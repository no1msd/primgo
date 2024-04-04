package primo

import (
	"image"
	"image/color"

	"primgo/primo/roms"
)

type ROMType string

const (
	ROMTypeA ROMType = "a"
	ROMTypeB ROMType = "b"
	ROMTypeC ROMType = "c"
)

func (r ROMType) Validate() bool {
	return map[ROMType]bool{
		ROMTypeA: true,
		ROMTypeB: true,
		ROMTypeC: true,
	}[r]
}

type ROMLabel string

const (
	ROMLabelINBYTE   ROMLabel = "inbyte"
	ROMLabelRDHEAD   ROMLabel = "rdhead"
	ROMLabelRDSYN    ROMLabel = "rdsyn"
	ROMLabelGOMBM    ROMLabel = "gombm"
	ROMLabelINIT     ROMLabel = "init"
	ROMLabelRESET    ROMLabel = "reset"
	ROMLabelNMIStuck ROMLabel = "nmi_stuck"
)

type ScreenPage string

const (
	ScreenPagePrimary   ScreenPage = "primary"
	ScreenPageSecondary ScreenPage = "secondary"
)

func (s ScreenPage) StartAddress() uint16 {
	switch s {
	case ScreenPagePrimary:
		return 0xE000
	case ScreenPageSecondary:
		return 0xC000
	}
	return 0
}

func (s ScreenPage) EndAddress() uint16 {
	switch s {
	case ScreenPagePrimary:
		return 0xFFFF
	case ScreenPageSecondary:
		return 0xDFFF
	}
	return 0
}

type coloringMode uint8

const (
	coloringMode4x4 coloringMode = 0
	coloringMode6x6 coloringMode = 2
	coloringMode6x9 coloringMode = 6
)

func (c coloringMode) size() image.Point {
	switch c {
	case coloringMode4x4:
		return image.Point{X: 4, Y: 4}
	case coloringMode6x6:
		return image.Point{X: 6, Y: 6}
	case coloringMode6x9:
		return image.Point{X: 6, Y: 9}
	}
	return image.Point{X: 4, Y: 4}
}

type Memory struct {
	ROMType      ROMType
	data         [0x10000]byte
	protected    uint16
	romLabelAdrs map[ROMLabel]map[ROMType]uint16
}

func NewMemory(romType ROMType) *Memory {
	var romData []byte
	switch romType {
	case ROMTypeA:
		romData = roms.A64
	case ROMTypeB:
		romData = roms.B64
	case ROMTypeC:
		romData = roms.C64
	}

	mem := &Memory{
		ROMType:   romType,
		protected: uint16(len(romData)),
		romLabelAdrs: map[ROMLabel]map[ROMType]uint16{
			ROMLabelINBYTE:   {ROMTypeA: 0x3CAB, ROMTypeB: 0x3CAB, ROMTypeC: 0x0DCC},
			ROMLabelRDHEAD:   {ROMTypeA: 0x3B36, ROMTypeB: 0x3B36, ROMTypeC: 0x0C58},
			ROMLabelRDSYN:    {ROMTypeA: 0x3C75, ROMTypeB: 0x3C75, ROMTypeC: 0x0D96},
			ROMLabelGOMBM:    {ROMTypeA: 0x3921, ROMTypeB: 0x3921},
			ROMLabelINIT:     {ROMTypeA: 0x3178, ROMTypeB: 0x3178, ROMTypeC: 0x00C9},
			ROMLabelRESET:    {ROMTypeA: 0x316A, ROMTypeB: 0x316A},
			ROMLabelNMIStuck: {ROMTypeC: 0x3e7f},
		},
	}
	copy(mem.data[:], romData)

	// patch GOMBM subroutine to reduce the num of repeated reads needed to register a keypress
	if romType != ROMTypeC {
		mem.data[mem.ROMLabelAddress(ROMLabelGOMBM)+3] = 48
	}

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

// decodeRGB will convert from the 1 byte Primo RGB information to 4 byte RGBA.
func decodeRGB(v uint8) color.RGBA {
	const redBitmask = 0xe0
	const greenBitmask = 0x1c
	const blueBitmask = 0x03

	return color.RGBA{
		R: v & redBitmask >> 5 * (0xff / (redBitmask >> 5)),
		G: v & greenBitmask >> 2 * (0xff / (greenBitmask >> 2)),
		B: v & blueBitmask * (0xff / blueBitmask),
		A: 0xff,
	}
}

func (m *Memory) coloringMode(screenPage ScreenPage) coloringMode {
	return coloringMode(m.Get(screenPage.StartAddress()))
}

func (m *Memory) activePalette(screenPage ScreenPage) uint16 {
	const palette1Bitmask = 0x01
	const palette2Bitmask = 0x02
	const palette3Bitmask = 0x04

	paletteIndex := m.Get(screenPage.StartAddress() + 1)
	if paletteIndex&palette1Bitmask != 0 {
		return 1
	}
	if paletteIndex&palette2Bitmask != 0 {
		return 2
	}
	if paletteIndex&palette3Bitmask != 0 {
		return 3
	}

	return 1
}

func (m *Memory) ScreenResolution(screenPage ScreenPage) image.Point {
	// the monochrome version always uses the same resolution
	if m.ROMType != ROMTypeC {
		return image.Point{X: 256, Y: 192}
	}

	// in the color version resolution depends on the size of the colorable chunks
	if m.coloringMode(screenPage) == coloringMode4x4 {
		return image.Point{X: 256, Y: 200}
	}

	return image.Point{X: 256, Y: 216}
}

func (m *Memory) monochromeColors(on bool) color.RGBA {
	// for the monochrome version we just use arbitrary black and white colors
	if on {
		return color.RGBA{R: 0xec, G: 0xec, B: 0xec, A: 0xff}
	}
	return color.RGBA{R: 0x18, G: 0x18, B: 0x18, A: 0xff}
}

func (m *Memory) pixelColorIndex(screenPage ScreenPage, row, col int) uint16 {
	chunkSize := m.coloringMode(screenPage).size()
	start := screenPage.StartAddress()
	chunkRow := row / chunkSize.Y
	chunkCol := col / chunkSize.X
	useUpper4bits := chunkCol%2 == 0
	offset := 0

	// 6 wide color chunk information starts with a 4 pixel wide first chunk
	if chunkSize.X == 6 {
		chunkCol = (col - 4) / chunkSize.X
		useUpper4bits = chunkCol%2 == 0
		if col >= 4 {
			offset = 1
		}
	}

	colorIndexAddr := start + 4*32 + uint16(chunkRow)*32 + uint16(chunkCol)/2 + uint16(offset)
	colorIndex := uint16(m.Get(colorIndexAddr))

	// a color index is 4 bit so we have to split this byte into two
	if useUpper4bits {
		return colorIndex >> 4
	}
	return colorIndex & 0x0F
}

func (m *Memory) pixelColor(screenPage ScreenPage, row, col int, on bool) color.RGBA {
	if m.ROMType != ROMTypeC {
		return m.monochromeColors(on)
	}

	colorIndex := m.pixelColorIndex(screenPage, row, col)
	colorAddr := screenPage.StartAddress() + m.activePalette(screenPage)*32 + colorIndex

	// lower 16 bytes are the background colors, the upper 16 bytes are the foreground colors
	if on {
		colorAddr += 16
	}

	return decodeRGB(m.Get(colorAddr))
}

func (m *Memory) GetRGBAScreenData(screenPage ScreenPage) []byte {
	screenSize := m.ScreenResolution(screenPage)
	pixels := make([]byte, (screenSize.X*screenSize.Y)*4)

	end := int(screenPage.EndAddress())
	px := 0
	for addr := end - screenSize.X*screenSize.Y/8 + 1; addr <= end; addr++ {
		b := m.Get(uint16(addr))

		// each byte contains data for 8 pixels
		for n := 7; n >= 0; n-- {
			row := px / screenSize.X
			col := px - row*screenSize.X
			pxColor := m.pixelColor(screenPage, row, col, (b>>n)&1 == 1)

			pixels[4*px] = pxColor.R
			pixels[4*px+1] = pxColor.G
			pixels[4*px+2] = pxColor.B
			pixels[4*px+3] = pxColor.A

			px++
		}
	}

	return pixels
}

func (m *Memory) ROMLabelAddress(label ROMLabel) uint16 {
	return m.romLabelAdrs[label][m.ROMType]
}
