package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/exp/maps"
)

const (
	keyboardImageWidth  = 931
	keyboardImageHeight = 300
)

type Keyboard struct {
	Height int
	IsOpen bool

	res          Resources
	scroll       float64
	keys         map[ebiten.Key]image.Rectangle
	scale        float64
	screenBound  image.Rectangle
	tweens       Tweens
	clickHandler *ClickHandler[ebiten.Key]
}

//nolint:funlen
func NewKeyboard(res Resources) *Keyboard {
	keys := map[ebiten.Key]image.Rectangle{
		ebiten.KeyUp:    {Min: image.Point{X: 43, Y: 8}, Max: image.Point{X: 101, Y: 66}},
		ebiten.Key1:     {Min: image.Point{X: 101, Y: 8}, Max: image.Point{X: 159, Y: 66}},
		ebiten.Key2:     {Min: image.Point{X: 159, Y: 8}, Max: image.Point{X: 218, Y: 66}},
		ebiten.Key3:     {Min: image.Point{X: 218, Y: 8}, Max: image.Point{X: 276, Y: 66}},
		ebiten.Key4:     {Min: image.Point{X: 276, Y: 8}, Max: image.Point{X: 334, Y: 66}},
		ebiten.Key5:     {Min: image.Point{X: 334, Y: 8}, Max: image.Point{X: 392, Y: 66}},
		ebiten.Key6:     {Min: image.Point{X: 392, Y: 8}, Max: image.Point{X: 450, Y: 66}},
		ebiten.Key7:     {Min: image.Point{X: 450, Y: 8}, Max: image.Point{X: 509, Y: 66}},
		ebiten.Key8:     {Min: image.Point{X: 509, Y: 8}, Max: image.Point{X: 568, Y: 66}},
		ebiten.Key9:     {Min: image.Point{X: 568, Y: 8}, Max: image.Point{X: 626, Y: 66}},
		ebiten.Key0:     {Min: image.Point{X: 626, Y: 8}, Max: image.Point{X: 684, Y: 66}},
		ebiten.KeyMinus: {Min: image.Point{X: 684, Y: 8}, Max: image.Point{X: 742, Y: 66}},
		ebiten.KeyEqual: {Min: image.Point{X: 742, Y: 8}, Max: image.Point{X: 801, Y: 66}},
		ebiten.KeyLeft:  {Min: image.Point{X: 801, Y: 8}, Max: image.Point{X: 859, Y: 66}},
		ebiten.KeyTab:   {Min: image.Point{X: 859, Y: 8}, Max: image.Point{X: 916, Y: 66}},

		ebiten.KeyControl:     {Min: image.Point{X: 13, Y: 67}, Max: image.Point{X: 71, Y: 124}},
		ebiten.KeyDown:        {Min: image.Point{X: 71, Y: 67}, Max: image.Point{X: 129, Y: 124}},
		ebiten.KeyQ:           {Min: image.Point{X: 129, Y: 67}, Max: image.Point{X: 188, Y: 124}},
		ebiten.KeyW:           {Min: image.Point{X: 188, Y: 67}, Max: image.Point{X: 246, Y: 124}},
		ebiten.KeyE:           {Min: image.Point{X: 246, Y: 67}, Max: image.Point{X: 304, Y: 124}},
		ebiten.KeyR:           {Min: image.Point{X: 304, Y: 67}, Max: image.Point{X: 362, Y: 124}},
		ebiten.KeyT:           {Min: image.Point{X: 362, Y: 67}, Max: image.Point{X: 421, Y: 124}},
		ebiten.KeyY:           {Min: image.Point{X: 421, Y: 67}, Max: image.Point{X: 479, Y: 124}},
		ebiten.KeyU:           {Min: image.Point{X: 479, Y: 67}, Max: image.Point{X: 537, Y: 124}},
		ebiten.KeyI:           {Min: image.Point{X: 537, Y: 67}, Max: image.Point{X: 596, Y: 124}},
		ebiten.KeyO:           {Min: image.Point{X: 596, Y: 67}, Max: image.Point{X: 654, Y: 124}},
		ebiten.KeyP:           {Min: image.Point{X: 654, Y: 67}, Max: image.Point{X: 712, Y: 124}},
		ebiten.KeyBracketLeft: {Min: image.Point{X: 712, Y: 67}, Max: image.Point{X: 770, Y: 124}},
		ebiten.KeyRight:       {Min: image.Point{X: 770, Y: 67}, Max: image.Point{X: 829, Y: 124}},
		ebiten.KeyEnter:       {Min: image.Point{X: 829, Y: 67}, Max: image.Point{X: 916, Y: 124}},

		ebiten.KeyCapsLock:     {Min: image.Point{X: 13, Y: 124}, Max: image.Point{X: 86, Y: 182}},
		ebiten.KeyA:            {Min: image.Point{X: 86, Y: 124}, Max: image.Point{X: 145, Y: 182}},
		ebiten.KeyS:            {Min: image.Point{X: 145, Y: 124}, Max: image.Point{X: 203, Y: 182}},
		ebiten.KeyD:            {Min: image.Point{X: 203, Y: 124}, Max: image.Point{X: 261, Y: 182}},
		ebiten.KeyF:            {Min: image.Point{X: 261, Y: 124}, Max: image.Point{X: 320, Y: 182}},
		ebiten.KeyG:            {Min: image.Point{X: 320, Y: 124}, Max: image.Point{X: 378, Y: 182}},
		ebiten.KeyH:            {Min: image.Point{X: 378, Y: 124}, Max: image.Point{X: 436, Y: 182}},
		ebiten.KeyJ:            {Min: image.Point{X: 436, Y: 124}, Max: image.Point{X: 494, Y: 182}},
		ebiten.KeyK:            {Min: image.Point{X: 494, Y: 124}, Max: image.Point{X: 553, Y: 182}},
		ebiten.KeyL:            {Min: image.Point{X: 553, Y: 124}, Max: image.Point{X: 611, Y: 182}},
		ebiten.KeySemicolon:    {Min: image.Point{X: 611, Y: 124}, Max: image.Point{X: 670, Y: 182}},
		ebiten.KeyQuote:        {Min: image.Point{X: 670, Y: 124}, Max: image.Point{X: 728, Y: 182}},
		ebiten.KeyInsert:       {Min: image.Point{X: 728, Y: 124}, Max: image.Point{X: 786, Y: 182}},
		ebiten.KeyBracketRight: {Min: image.Point{X: 786, Y: 124}, Max: image.Point{X: 844, Y: 182}},
		ebiten.KeyHome:         {Min: image.Point{X: 844, Y: 124}, Max: image.Point{X: 916, Y: 182}},

		ebiten.KeyShift:      {Min: image.Point{X: 15, Y: 182}, Max: image.Point{X: 111, Y: 240}},
		ebiten.KeyZ:          {Min: image.Point{X: 111, Y: 182}, Max: image.Point{X: 174, Y: 240}},
		ebiten.KeyX:          {Min: image.Point{X: 174, Y: 182}, Max: image.Point{X: 233, Y: 240}},
		ebiten.KeyC:          {Min: image.Point{X: 233, Y: 182}, Max: image.Point{X: 291, Y: 240}},
		ebiten.KeyV:          {Min: image.Point{X: 291, Y: 182}, Max: image.Point{X: 349, Y: 240}},
		ebiten.KeyB:          {Min: image.Point{X: 349, Y: 182}, Max: image.Point{X: 407, Y: 240}},
		ebiten.KeyN:          {Min: image.Point{X: 407, Y: 182}, Max: image.Point{X: 466, Y: 240}},
		ebiten.KeyM:          {Min: image.Point{X: 466, Y: 182}, Max: image.Point{X: 524, Y: 240}},
		ebiten.KeyComma:      {Min: image.Point{X: 524, Y: 182}, Max: image.Point{X: 583, Y: 240}},
		ebiten.KeyPeriod:     {Min: image.Point{X: 583, Y: 182}, Max: image.Point{X: 641, Y: 240}},
		ebiten.KeySlash:      {Min: image.Point{X: 641, Y: 182}, Max: image.Point{X: 699, Y: 240}},
		ebiten.KeyDelete:     {Min: image.Point{X: 699, Y: 182}, Max: image.Point{X: 757, Y: 240}},
		ebiten.KeyEnd:        {Min: image.Point{X: 757, Y: 182}, Max: image.Point{X: 819, Y: 240}},
		ebiten.KeyShiftRight: {Min: image.Point{X: 819, Y: 182}, Max: image.Point{X: 913, Y: 240}},

		ebiten.KeySpace: {Min: image.Point{X: 247, Y: 241}, Max: image.Point{X: 685, Y: 299}},

		ebiten.KeyF1: {Min: image.Point{X: 830, Y: 252}, Max: image.Point{X: 915, Y: 284}},
	}

	keyboard := &Keyboard{res: res, scroll: 1.0, keys: keys}
	keyboard.clickHandler = NewClickHandler(maps.Keys(keys), keyboard.boundingRectangleForKey)

	return keyboard
}

func (k *Keyboard) Open() {
	k.tweens.CancelAll()
	k.tweens.Add(NewTween(&k.scroll, 0.0, animLength))
	k.IsOpen = true
}

func (k *Keyboard) Close() {
	k.tweens.CancelAll()
	k.tweens.Add(NewTween(&k.scroll, 1.0, animLength))
	k.IsOpen = false
}

func (k *Keyboard) boundingRectangleForKey(key ebiten.Key) image.Rectangle {
	topLeft := image.Point{
		X: k.screenBound.Dx()/2 - int((keyboardImageWidth*k.scale)/2.0),
		Y: k.screenBound.Dy() - statusBarHeight - int(keyboardImageHeight*k.scale),
	}

	return image.Rectangle{
		Min: topLeft.Add(image.Point{
			X: int(float64(k.keys[key].Min.X) * k.scale),
			Y: int(float64(k.keys[key].Min.Y) * k.scale),
		}),
		Max: topLeft.Add(image.Point{
			X: int(float64(k.keys[key].Max.X) * k.scale),
			Y: int(float64(k.keys[key].Max.Y) * k.scale),
		}),
	}
}

func (k *Keyboard) Update(ignoreInput *bool) {
	k.tweens.Update()
	if !*ignoreInput && k.IsOpen {
		k.clickHandler.Update()
	}
}

func (k *Keyboard) Layout(w, h int) {
	k.screenBound = image.Rectangle{Max: image.Point{X: w, Y: h}}
}

func (k *Keyboard) Draw(screen *ebiten.Image) {
	targetSize := DrawImage(screen, k.res.keyboard, DrawImageOptions{
		TargetRect: image.Rectangle{
			Max: k.screenBound.Size().Sub(image.Point{Y: statusBarHeight}),
		},
		Filter:                 ebiten.FilterLinear,
		ScaleType:              ScaleTypeKeepAspectRatio,
		HorizontalAlign:        HorizontalAlignCenter,
		VerticalAlign:          VerticalAlignBottom,
		MaxSize:                image.Point{Y: keyboardImageHeight},
		ProportionalTranslateY: k.scroll,
	})
	k.Height = int(float64(targetSize.Y) * (1.0 - k.scroll))
	k.scale = float64(targetSize.X) / keyboardImageWidth

	var keys []ebiten.Key
	for _, key := range k.clickHandler.AppendAllHover(keys) {
		bound := k.boundingRectangleForKey(key)
		vector.DrawFilledRect(
			screen,
			float32(bound.Min.X),
			float32(bound.Min.Y),
			float32(bound.Dx()),
			float32(bound.Dy()),
			color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0x80},
			false)
	}
}

func (k *Keyboard) AppendPressedKeys(keys []ebiten.Key) []ebiten.Key {
	return k.clickHandler.AppendAllPressed(keys)
}
