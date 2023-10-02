package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/exp/slices"
)

const (
	listBackgroundItemID = "{bg}"
	listRowHeight        = 48
	listWidth            = 180
)

type Boundable interface {
	BoundingRectangle() image.Rectangle
}

type ItemInfo struct {
	Label     string
	ID        string
	Highlight bool
}

type PopupList struct {
	IsOpen  bool
	OnClick func(id string)

	items          []ItemInfo
	selectedItem   string
	anchor         Boundable
	positionOffset float64
	screenSize     image.Point
	res            Resources
	tweens         Tweens
	clickHandler   *ClickHandler[string]
}

func NewPopupList(items []ItemInfo, anchor Boundable, res Resources) *PopupList {
	popupList := &PopupList{
		items:          items,
		anchor:         anchor,
		positionOffset: float64((len(items) + 1) * listRowHeight),
		res:            res,
	}

	itemIDs := make([]string, 0, len(items))
	for _, item := range items {
		itemIDs = append(itemIDs, item.ID)
	}

	popupList.clickHandler = NewClickHandler(
		append(itemIDs, listBackgroundItemID), popupList.boundingRectangleForItem)
	popupList.clickHandler.OnReleased = popupList.onReleased

	return popupList
}

func (p *PopupList) onReleased(id string) {
	if p.OnClick != nil && id != listBackgroundItemID {
		p.OnClick(id)
		p.selectedItem = id
	}

	if id == listBackgroundItemID {
		p.Close()
	}
}

func (p *PopupList) Open() {
	p.tweens.CancelAll()
	p.tweens.Add(NewTween(&p.positionOffset, 0, animLength))
	p.IsOpen = true
}

func (p *PopupList) Close() {
	p.tweens.CancelAll()
	p.tweens.Add(NewTween(&p.positionOffset, float64((len(p.items)+1)*listRowHeight), animLength))
	p.IsOpen = false
}

func (p PopupList) drawItem(screen *ebiten.Image, n int) {
	bound := p.boundingRectangle()

	if p.items[n].Highlight {
		vector.StrokeLine(
			screen,
			float32(bound.Min.X), float32(bound.Min.Y+n*listRowHeight),
			float32(bound.Max.X), float32(bound.Min.Y+n*listRowHeight),
			1,
			color.RGBA{R: 0x3f, G: 0x3f, B: 0x3f, A: 0xff},
			false)
	}

	c := color.RGBA{0x97, 0x97, 0x97, 0xff}
	if p.clickHandler.Hover(p.items[n].ID) {
		c = color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
	text.Draw(
		screen,
		p.items[n].Label,
		p.res.font,
		bound.Min.X+12,
		bound.Min.Y+n*listRowHeight+28,
		c)

	if p.items[n].Label == p.selectedItem {
		vector.DrawFilledCircle(
			screen,
			float32(bound.Max.X-13),
			float32(bound.Min.Y+n*listRowHeight+23),
			4,
			c,
			true)
	}
}

func (p PopupList) Draw(screen *ebiten.Image) {
	bound := p.boundingRectangle()

	vector.DrawFilledRect(
		screen,
		float32(bound.Min.X), float32(bound.Min.Y),
		float32(bound.Dx()), float32(bound.Dy()),
		color.RGBA{R: 0x3f, G: 0x3f, B: 0x3f, A: 0xff},
		false)

	vector.DrawFilledRect(
		screen,
		float32(bound.Min.X)+1, float32(bound.Min.Y)+1,
		float32(bound.Dx())-2, float32(bound.Dy())-2,
		color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff},
		false)

	for n := range p.items {
		p.drawItem(screen, n)
	}
}

func (p PopupList) position() image.Point {
	return p.anchor.BoundingRectangle().Min.Add(p.anchor.BoundingRectangle().Size().Div(2))
}

func (p *PopupList) boundingRectangle() image.Rectangle {
	return image.Rectangle{
		Min: p.position().
			Sub(image.Point{X: listWidth, Y: len(p.items) * listRowHeight}).
			Add(image.Point{X: 0, Y: int(p.positionOffset)}),
		Max: p.position().
			Add(image.Point{X: 0, Y: int(p.positionOffset)}),
	}
}

func (p *PopupList) boundingRectangleForItem(id string) image.Rectangle {
	if id == listBackgroundItemID {
		return image.Rectangle{Min: image.Point{}, Max: p.screenSize}
	}

	idx := len(p.items) - slices.IndexFunc(p.items, func(ii ItemInfo) bool { return ii.ID == id })
	return image.Rectangle{
		Min: p.position().Sub(image.Point{X: listWidth, Y: idx * listRowHeight}),
		Max: p.position().Sub(image.Point{X: 0, Y: (idx - 1) * listRowHeight}),
	}
}

func (p *PopupList) Update(ignoreInput *bool) {
	p.tweens.Update()

	if p.IsOpen && !*ignoreInput {
		p.clickHandler.Update()
		*ignoreInput = true
	}
}

func (p *PopupList) Layout(w, h int) {
	p.screenSize = image.Point{X: w, Y: h}
}
