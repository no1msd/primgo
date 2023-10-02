package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/exp/slices"
)

const (
	mouseTouchID = -1
)

type ClickHandler[PartID comparable] struct {
	OnPressed  func(partID PartID)
	OnReleased func(partID PartID)

	hover             map[PartID]map[ebiten.TouchID]struct{}
	pressed           map[PartID]map[ebiten.TouchID]struct{}
	allHover          []PartID
	allPressed        []PartID
	boundingRectangle func(partID PartID) image.Rectangle
}

type MonoClickHandler = ClickHandler[struct{}]

func NewClickHandler[PartID comparable](
	partIDs []PartID,
	boundingRectangle func(partID PartID) image.Rectangle,
) *ClickHandler[PartID] {
	hover := make(map[PartID]map[ebiten.TouchID]struct{}, 0)
	pressed := make(map[PartID]map[ebiten.TouchID]struct{}, 0)
	for _, partID := range partIDs {
		hover[partID] = make(map[ebiten.TouchID]struct{}, 0)
		pressed[partID] = make(map[ebiten.TouchID]struct{}, 0)
	}

	return &ClickHandler[PartID]{
		hover:             hover,
		pressed:           pressed,
		boundingRectangle: boundingRectangle,
	}
}

func NewMonoClickHandler(boundingRectangle func() image.Rectangle) *ClickHandler[struct{}] {
	return NewClickHandler([]struct{}{{}}, func(struct{}) image.Rectangle {
		return boundingRectangle()
	})
}

func (t *ClickHandler[PartID]) handleHover(x, y int, touchID ebiten.TouchID) {
	for key := range t.hover {
		if image.Rect(x, y, x+1, y+1).In(t.boundingRectangle(key)) {
			if !t.Hover(key) {
				t.allHover = append(t.allHover, key)
			}
			t.hover[key][touchID] = struct{}{}
		} else {
			delete(t.hover[key], touchID)
			if !t.Hover(key) {
				t.allHover = slices.DeleteFunc(t.allHover, func(pi PartID) bool { return pi == key })
			}
		}
	}
}

func (t *ClickHandler[PartID]) handlePressed(touchID ebiten.TouchID) {
	for key := range t.hover {
		if !t.Hover(key) {
			continue
		}
		if !t.Pressed(key) {
			t.allPressed = append(t.allPressed, key)
		}
		t.pressed[key][touchID] = struct{}{}
		if t.OnPressed != nil {
			t.OnPressed(key)
		}
	}
}

func (t *ClickHandler[PartID]) handleReleased(touchID ebiten.TouchID) {
	if touchID != mouseTouchID {
		t.handleHover(-1, -1, touchID)
	}

	for key := range t.pressed {
		if !t.Pressed(key) {
			continue
		}
		delete(t.pressed[key], touchID)
		if !t.Pressed(key) {
			t.allPressed = slices.DeleteFunc(t.allPressed, func(pi PartID) bool { return pi == key })
		}
		if t.OnReleased != nil {
			t.OnReleased(key)
		}
	}
}

func (t *ClickHandler[PartID]) Update() {
	var touches []ebiten.TouchID

	mouseX, mouseY := ebiten.CursorPosition()
	t.handleHover(mouseX, mouseY, mouseTouchID)
	for _, touchID := range ebiten.AppendTouchIDs(touches[:0]) {
		x, y := ebiten.TouchPosition(touchID)
		t.handleHover(x, y, touchID)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		t.handlePressed(mouseTouchID)
	}
	for _, touchID := range inpututil.AppendJustPressedTouchIDs(touches[:0]) {
		t.handlePressed(touchID)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton0) {
		t.handleReleased(mouseTouchID)
	}
	for _, touchID := range inpututil.AppendJustReleasedTouchIDs(touches[:0]) {
		t.handleReleased(touchID)
	}
}

func (t ClickHandler[PartID]) AppendAllPressed(v []PartID) []PartID {
	return append(v, t.allPressed...)
}

func (t ClickHandler[PartID]) AppendAllHover(v []PartID) []PartID {
	return append(v, t.allHover...)
}

func (t ClickHandler[PartID]) AnyPressed() bool {
	return len(t.allPressed) > 0
}

func (t ClickHandler[PartID]) AnyHover() bool {
	return len(t.allHover) > 0
}

func (t ClickHandler[PartID]) Pressed(partID PartID) bool {
	return len(t.pressed[partID]) > 0
}

func (t ClickHandler[PartID]) Hover(partID PartID) bool {
	return len(t.hover[partID]) > 0
}
