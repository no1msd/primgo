package primo

import "golang.org/x/exp/slices"

const (
	inVblankBitmask      = 0x20
	inResetBitmask       = 0x02
	inKeyboardBitmask    = 0x01
	outNMIBitmask        = 0x80
	outSpeakerBitmask    = 0x10
	outScreenPageBitmask = 0x08
	outJoy1Bitmask       = 0x01
	outJoy2Bitmask       = 0x04
)

type IO struct {
	NMIEnabled   bool
	NMINext      bool
	PrimaryVideo bool
	Speaker      bool
	VBlank       bool
	Keys         []uint8
	Reset        bool
}

func NewIO() *IO {
	return &IO{
		NMIEnabled:   true,
		PrimaryVideo: true,
	}
}

func (i *IO) vblankBit() byte {
	if i.VBlank {
		return inVblankBitmask
	}
	return 0
}

func (i *IO) keyboardBit(address uint8) byte {
	if slices.Contains(i.Keys, address) {
		return inKeyboardBitmask
	}
	return 0
}

func (i *IO) resetBit() byte {
	if i.Reset {
		return inResetBitmask
	}
	return 0
}

func (i *IO) In(address uint8) uint8 {
	// 0x80-0xFF: Unused
	if address > 0x7f {
		return 0
	}

	// 0x40-0x7F: IN-2
	if address > 0x3f {
		return outJoy1Bitmask | outJoy2Bitmask
	}

	// 0x00-0x3F: IN-1
	return i.vblankBit() | i.keyboardBit(address) | i.resetBit()
}

func (i *IO) Out(address, b uint8) {
	// 0x80-0xFF: Unused
	if address > 0x7f {
		return
	}

	// 0x40-0x7F: OUT-2
	if address > 0x3f {
		return
	}

	// 0x00-0x3F: OUT-1
	i.NMIEnabled = b&outNMIBitmask != 0
	i.Speaker = b&outSpeakerBitmask != 0
	i.PrimaryVideo = b&outScreenPageBitmask != 0
}

func (i *IO) CheckNMI() bool {
	if !i.NMIEnabled || !i.NMINext {
		return false
	}
	i.NMINext = false
	return true
}
