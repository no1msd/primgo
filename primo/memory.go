package primo

import (
	"image/color"

	"primgo/primo/roms"
)

type ScreenType string

const (
	memorySize           = 0x10000
	primaryScreenStart   = 0xE800
	primaryScreenEnd     = 0xFFFF
	secondaryScreenStart = 0xC800
	secondaryScreenEnd   = 0xDFFF
)

type Memory struct {
	data      [memorySize]byte
	protected uint16
}

func NewMemory(keyboardRepeat byte) *Memory {
	romData := roms.A64

	// patch GOMBM subroutine to reduce the num of repeated reads needed to register a keypress
	romData[0x3924] = keyboardRepeat

	mem := &Memory{
		protected: uint16(len(romData)),
	}
	copy(mem.data[:], romData)
	return mem
}

func (m *Memory) Get(address uint16) uint8 {
	return m.data[address]
}

func (m *Memory) Set(address uint16, b uint8) {
	if address < m.protected {
		return
	}
	m.data[address] = b
}

func (m *Memory) GetRGBAScreenData(primaryScreen bool, bg, fg color.RGBA) []byte {
	pix := make([]byte, (primaryScreenEnd-primaryScreenStart+1)*8*4)

	start := secondaryScreenStart
	end := secondaryScreenEnd
	if primaryScreen {
		start = primaryScreenStart
		end = primaryScreenEnd
	}

	px := 0
	for addr := start; addr <= end; addr++ {
		b := m.Get(uint16(addr))
		// each byte contains data for 8 pixels
		for n := 7; n >= 0; n-- {
			pxColor := bg
			if (b>>n)&1 == 1 {
				pxColor = fg
			}

			pix[4*px] = pxColor.R
			pix[4*px+1] = pxColor.G
			pix[4*px+2] = pxColor.B
			pix[4*px+3] = pxColor.A

			px++
		}
	}

	return pix
}
