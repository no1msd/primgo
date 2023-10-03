package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/koron-go/z80"
	"golang.org/x/exp/slices"

	"primgo/primo"
	"primgo/ui"
)

const (
	screenWidth         = 256
	screenHeight        = 192
	tickPerSec          = 50
	vblankLength        = 0.0016
	keyboardRepeat      = 48
	audioBufferSizeInMS = 100
	sampleRate          = 44100
)

type Emulator struct {
	primoScreen *ebiten.Image
	bgColor     color.RGBA
	fgColor     color.RGBA

	memory *primo.Memory
	io     *primo.IO
	tape   *primo.TapePlayer
	audio  *primo.AudioBuffer
	cpu    *z80.CPU

	lastSoundSample float64
	ramInitialized  bool

	ui *ui.UI

	freqCounter    int
	freqCountStart int64

	keyMappings ui.KeyMappings
}

func NewEmulator() *Emulator {
	mem := primo.NewMemory(keyboardRepeat)
	io := primo.NewIO()
	cpu := z80.Build(z80.WithMemory(mem), z80.WithIO(io), z80.WithNMI(io))
	tapePlayer := primo.NewTapePlayer()
	audioBuffer := primo.NewAudioBuffer(sampleRate)

	audioPlayer, _ := audio.NewContext(sampleRate).NewPlayer(audioBuffer)
	audioPlayer.SetBufferSize(audioBufferSizeInMS * time.Millisecond)
	audioPlayer.Play()

	emuUI := ui.New(ui.NewResources(), screenWidth, screenHeight)
	emuUI.OnTapeChange = func(data []byte) {
		tapePlayer.ChangeTape(data)
	}

	return &Emulator{
		primoScreen: ebiten.NewImage(screenWidth, screenHeight),
		bgColor:     color.RGBA{R: 0x18, G: 0x18, B: 0x18, A: 0xff},
		fgColor:     color.RGBA{R: 0xec, G: 0xec, B: 0xec, A: 0xff},
		memory:      mem,
		io:          io,
		tape:        tapePlayer,
		audio:       audioBuffer,
		cpu:         cpu,
		ui:          emuUI,
		keyMappings: ui.GetKeyMappings(),
	}
}

func (e *Emulator) hardReset() {
	e.memory = primo.NewMemory(keyboardRepeat)
	e.io = primo.NewIO()
	e.cpu = z80.Build(z80.WithMemory(e.memory), z80.WithIO(e.io), z80.WithNMI(e.io))
	e.ramInitialized = false
	e.tape.Reset()
}

// patchPTPLoad applies runtime ROM patches to load data from a PTP file instead of the tape
// recorder IO ports.
func (e *Emulator) patchPTPLoad() {
	// skip sync reading in RDSYN subroutine
	if e.cpu.PC == 0x3C76 {
		e.cpu.PC = 0x3C7D
	}

	// overwrite INBYTE subroutine
	if e.cpu.PC == 0x3CAB {
		nextByte := e.tape.NextByte()        // read next byte from PTP
		e.cpu.DE.Hi = nextByte + e.cpu.DE.Hi // store checksum in D register
		e.cpu.AF.Hi = nextByte               // store byte in A register
		e.cpu.PC = 0x3CB8                    // jump to RET in original subroutine
	}

	// skip cassette handling in RDHEAD subroutine
	if e.cpu.PC == 0x3B3F {
		e.cpu.PC = 0x3BAD
	}
}

// sampleAudio appends the current state of the speaker output to the audio stream's buffer,
// taking the sample rate into account.
func (e *Emulator) sampleAudio() {
	if e.ui.Muted {
		return
	}

	e.lastSoundSample += float64(e.cpu.LastOpCycles)
	sampleCycles := float64(e.ui.ClockSpeed) / sampleRate
	if e.lastSoundSample > sampleCycles {
		e.lastSoundSample -= sampleCycles
		e.audio.AddSample(e.io.Speaker)
	}
}

func (e *Emulator) updateKeyboardInput() {
	var keys []ebiten.Key
	keys = inpututil.AppendPressedKeys(keys)
	keys = e.ui.AppendPressedKeys(keys)

	e.io.Keys = e.keyMappings.Translate(keys)
	e.io.Reset = slices.Contains(keys, ebiten.KeyF1)

	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && slices.Contains(keys, ebiten.KeyControl) {
		e.hardReset()
	}
}

func (e *Emulator) updateFreqCounter() {
	now := time.Now().UnixMilli()
	if now-e.freqCountStart > 1000 {
		e.ui.MesauredClock = fmt.Sprintf("%.2f MHz", float64(e.freqCounter)/1000000.0)
		e.freqCountStart = now
		e.freqCounter = 0
	}
}

func (e *Emulator) Update() error {
	e.updateKeyboardInput()

	vblankCycles := int(vblankLength * float64(e.ui.ClockSpeed))
	cyclesPerTick := int(e.ui.ClockSpeed) / tickPerSec

	if e.ramInitialized {
		e.io.NMINext = true
	}

	// Emulate 1/tickPerSec second worth of CPU time
	for i := 0; i < cyclesPerTick; i += e.cpu.LastOpCycles {
		// Simulated 50Hz VBlank signal
		e.io.VBlank = i < vblankCycles

		// INIT subroutine is called
		if e.cpu.PC == 0x3178 {
			e.ramInitialized = true
		}

		// RESET subroutine is called
		if e.cpu.PC == 0x316A {
			e.cpu.InNMI = false // we have to manually reset the CPU's NMI state
		}

		e.patchPTPLoad()
		e.sampleAudio()

		// execute a single instruction
		e.cpu.Step()

		e.freqCounter += e.cpu.LastOpCycles
	}

	e.ui.Update()
	e.updateFreqCounter()

	return nil
}

func (e *Emulator) Draw(screen *ebiten.Image) {
	e.primoScreen.WritePixels(e.memory.GetRGBAScreenData(e.io.PrimaryVideo, e.bgColor, e.fgColor))
	e.ui.Draw(screen, e.primoScreen)
}

func (e *Emulator) Layout(w, h int) (int, int) {
	e.ui.Layout(w, h)
	return w, h
}

func main() {
	ebiten.SetWindowSize(screenWidth*3, screenHeight*3+48)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("PrimGO")
	ebiten.SetWindowIcon([]image.Image{
		ui.LoadPNGAsset("assets/icon512.png"),
		ui.LoadPNGAsset("assets/icon256.png"),
		ui.LoadPNGAsset("assets/icon128.png"),
		ui.LoadPNGAsset("assets/icon64.png"),
		ui.LoadPNGAsset("assets/icon48.png"),
		ui.LoadPNGAsset("assets/icon32.png"),
	})
	ebiten.SetTPS(tickPerSec)

	emu := NewEmulator()

	if err := ebiten.RunGame(emu); err != nil {
		log.Fatal(err)
	}
}
