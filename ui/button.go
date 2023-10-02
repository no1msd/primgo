package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type ButtonAlign string

const (
	ButtonAlignBottomLeft  ButtonAlign = "bottomleft"
	ButtonAlignBottomRight ButtonAlign = "bottomright"
	ButtonAlignTopRight    ButtonAlign = "topright"

	buttonMargin     = 12
	buttonTouchInset = -10
)

type Button struct {
	OnPressed  func()
	OnReleased func()
	Icon       *ebiten.Image

	align        ButtonAlign
	position     int
	clickHandler *MonoClickHandler
	screenSize   image.Point
}

func NewIconButton(icon *ebiten.Image, align ButtonAlign, position int) *Button {
	button := &Button{
		Icon:     icon,
		align:    align,
		position: position,
	}

	button.clickHandler = NewMonoClickHandler(func() image.Rectangle {
		return button.BoundingRectangle().Inset(buttonTouchInset)
	})
	button.clickHandler.OnPressed = button.onPressed
	button.clickHandler.OnReleased = button.onReleased

	return button
}

func (b Button) Draw(screen *ebiten.Image) {
	colorScale := colorm.ColorM{}
	if !b.clickHandler.AnyPressed() && !b.clickHandler.AnyHover() {
		if b.align == ButtonAlignTopRight {
			colorScale.Scale(1.0, 1.0, 1.0, 0.0)
		} else {
			colorScale.Scale(1.0, 1.0, 1.0, 0.5)
		}
	} else if !b.clickHandler.AnyPressed() && b.clickHandler.AnyHover() {
		colorScale.Scale(1.0, 1.0, 1.0, 0.75)
	}

	DrawImage(screen, b.Icon, DrawImageOptions{
		ScaleType:  ScaleTypeKeepAspectRatio,
		TargetRect: b.BoundingRectangle(),
		ColorScale: colorScale,
	})
}

func (b *Button) Update(ignoreInput *bool) {
	if !*ignoreInput {
		b.clickHandler.Update()
	}
}

func (b *Button) Layout(w, h int) {
	b.screenSize = image.Point{X: w, Y: h}
}

func (b *Button) BoundingRectangle() image.Rectangle {
	pos := b.screenSize.Sub(image.Point{
		X: ((buttonMargin*2)+b.Icon.Bounds().Dx())*(b.position+1) - buttonMargin,
		Y: statusBarHeight - (statusBarHeight-b.Icon.Bounds().Dy())/2,
	})

	if b.align == ButtonAlignTopRight {
		pos.Y = buttonMargin
	}

	if b.align == ButtonAlignBottomLeft {
		pos.X = ((buttonMargin*2)+b.Icon.Bounds().Dx())*b.position + buttonMargin
	}

	return image.Rectangle{
		Min: pos,
		Max: pos.Add(b.Icon.Bounds().Size()),
	}
}

func (b *Button) onPressed(struct{}) {
	if b.OnPressed != nil {
		b.OnPressed()
	}
}

func (b *Button) onReleased(struct{}) {
	if b.OnReleased != nil {
		b.OnReleased()
	}
}
