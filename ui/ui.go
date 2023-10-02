package ui

import (
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"

	"primgo/primo/tapes"
	"primgo/ui/dialog"
)

type ClockSpeed int

const (
	ClockSpeedNormal   = 2500000
	ClockSpeedSpectrum = 3500000
	ClockSpeedTurbo    = 3750000

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

type UI struct {
	Muted         bool
	ClockSpeed    ClockSpeed
	LoadedTape    string
	MesauredClock string
	OnTapeChange  func(data []byte)

	res             Resources
	displayStretch  bool
	upscaledScreens map[int]*ebiten.Image
	openedFileChan  chan *dialog.OpenedFile

	volumeButton   *Button
	tapeButton     *Button
	keyboardButton *Button
	freqButton     *Button
	displayButton  *Button
	keyboard       *Keyboard
	tapeList       *PopupList
}

func New(res Resources, screenWidth, screenHeight int) *UI {
	upscaledScreens := make(map[int]*ebiten.Image, maxWholeUpscale)
	for i := 1; i <= maxWholeUpscale; i++ {
		upscaledScreens[i-1] = ebiten.NewImage(screenWidth*i, screenHeight*i)
	}

	tapeButton := NewIconButton(res.tapeIconImage, ButtonAlignBottomRight, 2)

	ui := &UI{
		ClockSpeed:     ClockSpeedNormal,
		volumeButton:   NewIconButton(res.volumeIconImage, ButtonAlignBottomRight, 0),
		keyboardButton: NewIconButton(res.keyboardUpIconImage, ButtonAlignBottomRight, 1),
		tapeButton:     tapeButton,
		freqButton:     NewIconButton(res.cpu1IconImage, ButtonAlignBottomLeft, 0),
		displayButton:  NewIconButton(res.scale2IconImage, ButtonAlignTopRight, 0),
		keyboard:       NewKeyboard(res),
		tapeList: NewPopupList(
			[]ItemInfo{
				{Label: "emblema.ptp", ID: "emblema.ptp"},
				{Label: "betuk.ptp", ID: "betuk.ptp"},
				{Label: "rajzolo.ptp", ID: "rajzolo.ptp"},
				{Label: "kigyo.ptp", ID: "kigyo.ptp"},
				{Label: "othello.ptp", ID: "othello.ptp"},
				{Label: "hammm.ptp", ID: "hammm.ptp"},
				{Label: "himnusz.ptp", ID: "himnusz.ptp"},
				{Label: "foldrajz.ptp", ID: "foldrajz.ptp"},
				{Label: "Open PTP file", ID: openPTPItemID, Highlight: true},
			},
			tapeButton,
			res),
		res:             res,
		LoadedTape:      "[empty]",
		displayStretch:  true,
		upscaledScreens: upscaledScreens,
	}

	ui.registerCallbacks()

	return ui
}

func (s *UI) widgets() []Widget {
	return []Widget{
		s.tapeList,
		s.volumeButton,
		s.tapeButton,
		s.keyboardButton,
		s.freqButton,
		s.displayButton,
		s.keyboard,
	}
}

func (s *UI) registerCallbacks() {
	s.volumeButton.OnReleased = s.onVolumeClicked
	s.freqButton.OnReleased = s.onFreqClicked
	s.keyboardButton.OnReleased = s.onKeyboardClicked
	s.tapeButton.OnReleased = s.onTapeClicked
	s.displayButton.OnReleased = s.onDisplayClicked
	s.tapeList.OnClick = s.onTapeListClicked
}

func (s *UI) onDisplayClicked() {
	if s.displayStretch {
		s.displayButton.Icon = s.res.scale1IconImage
	} else {
		s.displayButton.Icon = s.res.scale2IconImage
	}
	s.displayStretch = !s.displayStretch
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

func (s *UI) onVolumeClicked() {
	if s.Muted {
		s.volumeButton.Icon = s.res.volumeIconImage
	} else {
		s.volumeButton.Icon = s.res.muteIconImage
	}
	s.Muted = !s.Muted
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

func (s *UI) onFreqClicked() {
	switch s.ClockSpeed {
	case ClockSpeedNormal:
		s.ClockSpeed = ClockSpeedSpectrum
		s.freqButton.Icon = s.res.cpu2IconImage
	case ClockSpeedSpectrum:
		s.ClockSpeed = ClockSpeedTurbo
		s.freqButton.Icon = s.res.cpu3IconImage
	case ClockSpeedTurbo:
		s.ClockSpeed = ClockSpeedNormal
		s.freqButton.Icon = s.res.cpu1IconImage
	}
}

func (s *UI) drawEmulatorScreen(screen, primoScreen *ebiten.Image) {
	// calculate the potential target bounds of the emulator screen in the window
	targetRect := image.Rectangle{
		Max: image.Point{
			X: screen.Bounds().Dx(),
			Y: screen.Bounds().Dy() - (statusBarHeight + s.keyboard.Height),
		},
	}

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
	scaleType := ScaleTypeKeepSize
	if s.displayStretch {
		scaleType = ScaleTypeKeepAspectRatio
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
	s.displayButton.Draw(screen)

	s.tapeList.Draw(screen)
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
