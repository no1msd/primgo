package ui

import "github.com/hajimehoshi/ebiten/v2"

type KeyMappings map[ebiten.Key]uint8

func (k KeyMappings) Translate(keys []ebiten.Key) []uint8 {
	ret := []uint8{}
	for _, key := range keys {
		if v, ok := k[key]; ok {
			ret = append(ret, v)
		}
	}
	return ret
}

//nolint:funlen
func GetKeyMappings() map[ebiten.Key]uint8 {
	return map[ebiten.Key]uint8{
		ebiten.KeyZ:            0x00, // Y
		ebiten.KeyUp:           0x01, // Up
		ebiten.KeyS:            0x02, // S
		ebiten.KeyShift:        0x03, // Shift
		ebiten.KeyShiftRight:   0x03, // Shift
		ebiten.KeyE:            0x04, // E
		ebiten.KeyCapsLock:     0x05, // Upper
		ebiten.KeyW:            0x06, // W
		ebiten.KeyControl:      0x07, // CTR
		ebiten.KeyD:            0x08, // D
		ebiten.Key3:            0x09, // 3 #
		ebiten.KeyX:            0x0a, // X
		ebiten.Key2:            0x0b, // 2 "
		ebiten.KeyQ:            0x0c, // Q
		ebiten.Key1:            0x0d, // 1 !
		ebiten.KeyA:            0x0e, // A
		ebiten.KeyDown:         0x0f, // Down
		ebiten.KeyC:            0x10, // C
		ebiten.KeyF:            0x12, // F
		ebiten.KeyR:            0x14, // R
		ebiten.KeyT:            0x16, // T
		ebiten.Key7:            0x17, // 7 /
		ebiten.KeyH:            0x18, // H
		ebiten.KeySpace:        0x19, // Space
		ebiten.KeyB:            0x1a, // B
		ebiten.Key6:            0x1b, // 6 &
		ebiten.KeyG:            0x1c, // G
		ebiten.Key5:            0x1d, // 5 %
		ebiten.KeyV:            0x1e, // V
		ebiten.Key4:            0x1f, // 4 $
		ebiten.KeyN:            0x20, // N
		ebiten.Key8:            0x21, // 8 (
		ebiten.KeyY:            0x22, // Z
		ebiten.KeyMinus:        0x23, // + ?
		ebiten.KeyU:            0x24, // U
		ebiten.Key0:            0x25, // O =
		ebiten.KeyJ:            0x26, // J
		ebiten.KeyDelete:       0x27, // > <
		ebiten.KeyL:            0x28, // L
		ebiten.KeySlash:        0x29, // - ⓘ
		ebiten.KeyK:            0x2a, // K
		ebiten.KeyPeriod:       0x2b, // . :
		ebiten.KeyM:            0x2c, // M
		ebiten.Key9:            0x2d, // 9 )
		ebiten.KeyI:            0x2e, // I
		ebiten.KeyComma:        0x2f, // , .
		ebiten.KeyEqual:        0x30, // Ü
		ebiten.KeyInsert:       0x31, // ' *
		ebiten.KeyP:            0x32, // P
		ebiten.KeyEnd:          0x33, // ú ű
		ebiten.KeyO:            0x34, // O
		ebiten.KeyHome:         0x35, // CLS
		ebiten.KeyEnter:        0x37, // Return
		ebiten.KeyBackspace:    0x39, // Left
		ebiten.KeyLeft:         0x39, // Left
		ebiten.KeySemicolon:    0x3a, // É
		ebiten.KeyBracketRight: 0x3b, // ó ő
		ebiten.KeyQuote:        0x3c, // Á
		ebiten.KeyRight:        0x3d, // Right
		ebiten.KeyBracketLeft:  0x3e, // Ö
		ebiten.KeyTab:          0x3f, // BRK
	}
}
