package main

import (
	"fmt"
	"image"
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
	tickPerSec          = 50
	vblankLength        = 0.0016
	audioBufferSizeInMS = 100
	sampleRate          = 44100
)

type Emulator struct {
	primoScreen *ebiten.Image

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
	emuUI := ui.New(ui.NewResources())

	mem := primo.NewMemory(emuUI.ROMType)
	io := primo.NewIO()
	cpu := z80.Build(z80.WithMemory(mem), z80.WithIO(io), z80.WithNMI(io))
	tapePlayer := primo.NewTapePlayer()
	audioBuffer := primo.NewAudioBuffer(sampleRate)

	audioPlayer, _ := audio.NewContext(sampleRate).NewPlayer(audioBuffer)
	audioPlayer.SetBufferSize(audioBufferSizeInMS * time.Millisecond)
	audioPlayer.Play()

	emu := &Emulator{
		memory:      mem,
		io:          io,
		tape:        tapePlayer,
		audio:       audioBuffer,
		cpu:         cpu,
		ui:          emuUI,
		keyMappings: ui.GetKeyMappings(),
	}

	emuUI.OnTapeChange = func(data []byte) {
		tapePlayer.ChangeTape(data)
	}

	emuUI.OnROMTypeChange = func(romType primo.ROMType) {
		emu.memory = primo.NewMemory(romType)
		emu.hardReset()
	}

	return emu
}

func (e *Emulator) hardReset() {
	e.io = primo.NewIO()
	e.cpu = z80.Build(z80.WithMemory(e.memory), z80.WithIO(e.io), z80.WithNMI(e.io))
	e.ramInitialized = false
	e.tape.Reset()
}

// patchPTPLoad applies runtime ROM patches to load data from a PTP file instead of the tape
// recorder IO ports.
func (e *Emulator) patchPTPLoad() {
	// skip sync reading in RDSYN subroutine
	if e.cpu.PC == e.memory.ROMLabelAddress(primo.ROMLabelRDSYN) {
		e.cpu.PC += 8
	}

	// overwrite INBYTE subroutine
	if e.cpu.PC == e.memory.ROMLabelAddress(primo.ROMLabelINBYTE) {
		nextByte := e.tape.NextByte()        // read next byte from PTP
		e.cpu.DE.Hi = nextByte + e.cpu.DE.Hi // store checksum in D register
		e.cpu.AF.Hi = nextByte               // store byte in A register
		e.cpu.PC += 13                       // jump to RET in original subroutine
	}

	// skip cassette handling in RDHEAD subroutine
	if e.cpu.PC == e.memory.ROMLabelAddress(primo.ROMLabelRDHEAD)+9 {
		e.cpu.PC += 110
	}
}

// patchStuckNMIHandler works around the issue of getting the execution stuck in the NMI handlers
// after a hard reset in the "C" version of the ROM.
func (e *Emulator) patchStuckNMIHandler() {
	if e.memory.ROMType != primo.ROMTypeC {
		return
	}

	if e.cpu.PC == e.memory.ROMLabelAddress(primo.ROMLabelNMIStuck) {
		e.cpu.PC++ // skip a jump that gets us stuck
	}
}

// patchStuckNMIFlag works around the issue of getting the CPU's InNMI flag stuck after a soft reset
// is executed in the "A" and "B" versions of the ROM.
func (e *Emulator) patchStuckNMIFlag() {
	if e.memory.ROMType != primo.ROMTypeA && e.memory.ROMType != primo.ROMTypeB {
		return
	}

	if e.cpu.PC == e.memory.ROMLabelAddress(primo.ROMLabelRESET) {
		e.cpu.InNMI = false // we have to manually reset the CPU's NMI state
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
		if e.cpu.PC == e.memory.ROMLabelAddress(primo.ROMLabelINIT) {
			e.ramInitialized = true
		}

		e.patchPTPLoad()
		e.patchStuckNMIHandler()
		e.patchStuckNMIFlag()
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
	screenPage := primo.ScreenPageSecondary
	if e.io.PrimaryVideo {
		screenPage = primo.ScreenPagePrimary
	}

	// ensure correct screen size
	desiredSize := e.memory.ScreenResolution(screenPage)
	if e.primoScreen == nil || e.primoScreen.Bounds().Size() != desiredSize {
		e.primoScreen = ebiten.NewImage(desiredSize.X, desiredSize.Y)
	}

	e.primoScreen.WritePixels(e.memory.GetRGBAScreenData(screenPage))
	e.ui.Draw(screen, e.primoScreen)
}

func (e *Emulator) Layout(w, h int) (int, int) {
	e.ui.Layout(w, h)
	return w, h
}

func main() {
	ebiten.SetWindowSize(768, 624)
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
