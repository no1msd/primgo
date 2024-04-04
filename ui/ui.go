package ui

import (
	"encoding/json"
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"

	"primgo/primo"
	"primgo/primo/tapes"
	"primgo/settings"
	"primgo/ui/dialog"
)

type ClockSpeed int

const (
	ClockSpeedNormal   = 2500000
	ClockSpeedSpectrum = 3500000
	ClockSpeedTurbo    = 3750000
)

func (c ClockSpeed) Validate() bool {
	return map[ClockSpeed]bool{
		ClockSpeedNormal:   true,
		ClockSpeedSpectrum: true,
		ClockSpeedTurbo:    true,
	}[c]
}

const (
	maxWholeUpscale = 8
	statusBarHeight = 48
	textMargin      = 14
	animLength      = 200 * time.Millisecond
	openPTPItemID   = "{ptp}"
)

type Widget interface {
	Draw(screen *ebiten.Image)
	Update(ignoreInput *bool)
	Layout(w, h int)
}

type primgoSettings struct {
	Muted          bool          `json:"muted"`
	WholeScaleOnly bool          `json:"whole_scale"`
	ClockSpeed     ClockSpeed    `json:"clock_speed"`
	ROMType        primo.ROMType `json:"rom_type"`
}

type UI struct {
	Muted           bool
	ClockSpeed      ClockSpeed
	ROMType         primo.ROMType
	LoadedTape      string
	MesauredClock   string
	OnTapeChange    func(data []byte)
	OnROMTypeChange func(romType primo.ROMType)

	res             Resources
	wholeScaleOnly  bool
	upscaledScreens map[int]*ebiten.Image
	openedFileChan  chan *dialog.OpenedFile

	volumeButton   *Button
	tapeButton     *Button
	keyboardButton *Button
	freqButton     *Button
	romButton      *Button
	displayButton  *Button
	keyboard       *Keyboard
	tapeList       *PopupList
	romList        *PopupList
}

func New(res Resources) *UI {
	upscaledScreens := make(map[int]*ebiten.Image, maxWholeUpscale)

	tapeButton := NewIconButton(res.tapeIconImage, ButtonAlignBottomRight, 2)
	romButton := NewIconButton(res.rom1IconImage, ButtonAlignBottomLeft, 0)

	ui := &UI{
		volumeButton:   NewIconButton(res.volumeIconImage, ButtonAlignBottomRight, 0),
		keyboardButton: NewIconButton(res.keyboardUpIconImage, ButtonAlignBottomRight, 1),
		tapeButton:     tapeButton,
		romButton:      romButton,
		freqButton:     NewIconButton(res.cpu1IconImage, ButtonAlignBottomLeft, 1),
		displayButton:  NewIconButton(res.scale2IconImage, ButtonAlignTopRight, 0),
		keyboard:       NewKeyboard(res),
		tapeList:       NewPopupList(tapeItems(), tapeButton, PopupAlignLeft, res),
		romList: NewPopupList(
			[]ItemInfo{
				{Label: "Reset to A64", ID: string(primo.ROMTypeA)},
				{Label: "Reset to B64", ID: string(primo.ROMTypeB)},
				{Label: "Reset to C64", ID: string(primo.ROMTypeC)},
			},
			romButton,
			PopupAlignRight,
			res),
		res:             res,
		LoadedTape:      "[empty]",
		upscaledScreens: upscaledScreens,
	}

	ui.registerCallbacks()
	ui.loadSettings()

	return ui
}

func tapeItems() []ItemInfo {
	return []ItemInfo{
		{Label: "raktaros.ptp", ID: "raktaros.ptp"},
		{Label: "emblema.ptp", ID: "emblema.ptp"},
		{Label: "betuk.ptp", ID: "betuk.ptp"},
		{Label: "rajzolo.ptp", ID: "rajzolo.ptp"},
		{Label: "kigyo.ptp", ID: "kigyo.ptp"},
		{Label: "othello.ptp", ID: "othello.ptp"},
		{Label: "hammm.ptp", ID: "hammm.ptp"},
		{Label: "himnusz.ptp", ID: "himnusz.ptp"},
		{Label: "foldrajz.ptp", ID: "foldrajz.ptp"},
		{Label: "Open PTP file", ID: openPTPItemID, Highlight: true},
	}
}

func (s *UI) loadSettings() {
	data, err := settings.Load()
	if err != nil {
		log.Printf("Error loading settings: %s\n", err.Error())
	}

	var ps primgoSettings
	// empty data is not an error, we just use the default values
	if data != "" {
		err = json.Unmarshal([]byte(data), &ps)
		if err != nil {
			log.Printf("Error unmarshalling settings: %s\n", err.Error())
		}
	}

	if !ps.ClockSpeed.Validate() {
		ps.ClockSpeed = ClockSpeedNormal
	}

	if !ps.ROMType.Validate() {
		ps.ROMType = primo.ROMTypeA
	}

	s.Muted = ps.Muted
	s.wholeScaleOnly = ps.WholeScaleOnly
	s.ClockSpeed = ps.ClockSpeed
	s.ROMType = ps.ROMType

	s.updateVolumeIcon()
	s.updateDisplayIcon()
	s.updateFreqIcon()
	s.updateROMIcon()
}

func (s *UI) saveSettings() {
	data, err := json.Marshal(primgoSettings{
		Muted:          s.Muted,
		WholeScaleOnly: s.wholeScaleOnly,
		ClockSpeed:     s.ClockSpeed,
		ROMType:        s.ROMType,
	})
	if err != nil {
		log.Printf("Error marshalling settings: %s\n", err.Error())
		return
	}
	err = settings.Save(string(data))
	if err != nil {
		log.Printf("Error saving settings: %s\n", err.Error())
	}
}

func (s *UI) widgets() []Widget {
	return []Widget{
		s.tapeList,
		s.romList,
		s.volumeButton,
		s.tapeButton,
		s.keyboardButton,
		s.romButton,
		s.freqButton,
		s.displayButton,
		s.keyboard,
	}
}

func (s *UI) registerCallbacks() {
	s.volumeButton.OnReleased = s.onVolumeClicked
	s.freqButton.OnReleased = s.onFreqClicked
	s.romButton.OnReleased = s.onROMClicked
	s.keyboardButton.OnReleased = s.onKeyboardClicked
	s.tapeButton.OnReleased = s.onTapeClicked
	s.displayButton.OnReleased = s.onDisplayClicked
	s.tapeList.OnClick = s.onTapeListClicked
	s.romList.OnClick = s.onROMListClicked
}

func (s *UI) updateDisplayIcon() {
	if s.wholeScaleOnly {
		s.displayButton.Icon = s.res.scale1IconImage
	} else {
		s.displayButton.Icon = s.res.scale2IconImage
	}
}

func (s *UI) onDisplayClicked() {
	s.wholeScaleOnly = !s.wholeScaleOnly
	s.updateDisplayIcon()
	s.saveSettings()
}

func (s *UI) onTapeListClicked(id string) {
	if id != openPTPItemID {
		if s.OnTapeChange != nil {
			s.OnTapeChange(tapes.ByName(id))
		}
		s.LoadedTape = id
		return
	}

	s.openedFileChan = dialog.BrowseFile()
}

func (s *UI) onROMListClicked(id string) {
	s.ROMType = primo.ROMType(id)
	if s.OnROMTypeChange != nil {
		s.OnROMTypeChange(s.ROMType)
	}
	s.updateROMIcon()
	s.saveSettings()
}

func (s *UI) updateVolumeIcon() {
	if s.Muted {
		s.volumeButton.Icon = s.res.muteIconImage
	} else {
		s.volumeButton.Icon = s.res.volumeIconImage
	}
}

func (s *UI) onVolumeClicked() {
	s.Muted = !s.Muted
	s.updateVolumeIcon()
	s.saveSettings()
}

func (s *UI) onKeyboardClicked() {
	if s.keyboard.IsOpen {
		s.keyboardButton.Icon = s.res.keyboardUpIconImage
		s.keyboard.Close()
	} else {
		s.keyboardButton.Icon = s.res.keyboardDownIconImage
		s.keyboard.Open()
	}
}

func (s *UI) onTapeClicked() {
	if !s.tapeList.IsOpen {
		s.tapeList.Open()
	}
}

func (s *UI) onROMClicked() {
	if !s.romList.IsOpen {
		s.romList.Open()
	}
}

func (s *UI) updateFreqIcon() {
	switch s.ClockSpeed {
	case ClockSpeedNormal:
		s.freqButton.Icon = s.res.cpu1IconImage
	case ClockSpeedSpectrum:
		s.freqButton.Icon = s.res.cpu2IconImage
	case ClockSpeedTurbo:
		s.freqButton.Icon = s.res.cpu3IconImage
	}
}

func (s *UI) updateROMIcon() {
	switch s.ROMType {
	case primo.ROMTypeA:
		s.romButton.Icon = s.res.rom1IconImage
	case primo.ROMTypeB:
		s.romButton.Icon = s.res.rom2IconImage
	case primo.ROMTypeC:
		s.romButton.Icon = s.res.rom3IconImage
	}
}

func (s *UI) onFreqClicked() {
	switch s.ClockSpeed {
	case ClockSpeedNormal:
		s.ClockSpeed = ClockSpeedSpectrum
	case ClockSpeedSpectrum:
		s.ClockSpeed = ClockSpeedTurbo
	case ClockSpeedTurbo:
		s.ClockSpeed = ClockSpeedNormal
	}
	s.updateFreqIcon()
	s.saveSettings()
}

func (s *UI) drawEmulatorScreen(screen, primoScreen *ebiten.Image) {
	// calculate the potential target bounds of the emulator screen in the window
	targetRect := image.Rectangle{
		Max: image.Point{
			X: screen.Bounds().Dx(),
			Y: screen.Bounds().Dy() - (statusBarHeight + s.keyboard.Height),
		},
	}

	// ensure upscaled screens have the correct resolution
	s.ensureUpscaledScreenSize(primoScreen)

	// upscale the raw image by the closest whole number to the target bounds for crispy pixels
	wholeScreenIdx := func(v float64) int {
		return int(math.Max(math.Min(math.Floor(v), maxWholeUpscale), 1)) - 1
	}
	idx := wholeScreenIdx(float64(targetRect.Bounds().Dx()) / float64(primoScreen.Bounds().Dx()))
	if s.upscaledScreens[idx].Bounds().Dy() > targetRect.Dy() {
		idx = wholeScreenIdx(float64(targetRect.Bounds().Dy()) / float64(primoScreen.Bounds().Dy()))
	}
	DrawImage(s.upscaledScreens[idx], primoScreen, DrawImageOptions{})

	// clear background
	screen.Fill(color.RGBA{0x1f, 0x1f, 0x1f, 0xff})

	// check if we want to scale to fractions or keep it crispy
	scaleType := ScaleTypeKeepAspectRatio
	if s.wholeScaleOnly {
		scaleType = ScaleTypeKeepSize
	}

	// finally draw the image to the target size with chosen scaling
	DrawImage(screen, s.upscaledScreens[idx], DrawImageOptions{
		ScaleType:       scaleType,
		HorizontalAlign: HorizontalAlignCenter,
		VerticalAlign:   VerticalAlignCenter,
		TargetRect:      targetRect,
		Filter:          ebiten.FilterLinear,
	})
}

func (s *UI) ensureUpscaledScreenSize(primoScreen *ebiten.Image) {
	for i := 1; i <= maxWholeUpscale; i++ {
		desiredSize := primoScreen.Bounds().Size().Mul(i)
		if s.upscaledScreens[i-1] != nil && s.upscaledScreens[i-1].Bounds().Size() == desiredSize {
			continue
		}
		s.upscaledScreens[i-1] = ebiten.NewImage(desiredSize.X, desiredSize.Y)
	}
}

func (s *UI) drawStatusBar(screen *ebiten.Image) {
	vector.DrawFilledRect(
		screen,
		0, float32(screen.Bounds().Dy()-statusBarHeight),
		float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()),
		color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff},
		false)

	vector.StrokeLine(
		screen,
		0, float32(screen.Bounds().Dy()-statusBarHeight),
		float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()-statusBarHeight),
		1,
		color.RGBA{R: 0x3f, G: 0x3f, B: 0x3f, A: 0xff},
		false)

	clockLabelBounds, _ := font.BoundString(s.res.font, s.MesauredClock)
	fontHeight := clockLabelBounds.Max.Y.Round() - clockLabelBounds.Min.Y.Round() - 2
	text.Draw(
		screen,
		s.MesauredClock,
		s.res.font,
		s.freqButton.BoundingRectangle().Max.X+textMargin,
		screen.Bounds().Max.Y-statusBarHeight/2+fontHeight/2,
		color.RGBA{0x97, 0x97, 0x97, 0xff})

	tapeLabelBounds, _ := font.BoundString(s.res.font, s.LoadedTape)
	tapeLabelWidth := tapeLabelBounds.Max.X.Round() - tapeLabelBounds.Min.X.Round()
	text.Draw(
		screen,
		s.LoadedTape,
		s.res.font,
		s.tapeButton.BoundingRectangle().Min.X-tapeLabelWidth-textMargin,
		screen.Bounds().Max.Y-statusBarHeight/2+fontHeight/2,
		color.RGBA{0x97, 0x97, 0x97, 0xff})
}

func (s *UI) Draw(screen, primoScreen *ebiten.Image) {
	s.drawEmulatorScreen(screen, primoScreen)

	s.keyboard.Draw(screen)

	s.drawStatusBar(screen)

	s.volumeButton.Draw(screen)
	s.tapeButton.Draw(screen)
	s.keyboardButton.Draw(screen)
	s.freqButton.Draw(screen)
	s.romButton.Draw(screen)
	s.displayButton.Draw(screen)

	s.tapeList.Draw(screen)
	s.romList.Draw(screen)
}

func (s *UI) Update() {
	ignoreInput := false
	for _, widget := range s.widgets() {
		widget.Update(&ignoreInput)
	}

	select {
	case openedFile := <-s.openedFileChan:
		if openedFile != nil {
			if s.OnTapeChange != nil {
				s.OnTapeChange(openedFile.Data)
			}
			s.LoadedTape = openedFile.Name
		}
	default:
	}
}

func (s UI) Layout(w, h int) {
	for _, widget := range s.widgets() {
		widget.Layout(w, h)
	}
}

func (s *UI) AppendPressedKeys(keys []ebiten.Key) []ebiten.Key {
	return s.keyboard.AppendPressedKeys(keys)
}
