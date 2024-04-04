package ui

import (
	"embed"
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed assets/*
var assetsFS embed.FS

type Resources struct {
	tapeIconImage         *ebiten.Image
	keyboardUpIconImage   *ebiten.Image
	keyboardDownIconImage *ebiten.Image
	volumeIconImage       *ebiten.Image
	muteIconImage         *ebiten.Image
	cpu1IconImage         *ebiten.Image
	cpu2IconImage         *ebiten.Image
	cpu3IconImage         *ebiten.Image
	rom1IconImage         *ebiten.Image
	rom2IconImage         *ebiten.Image
	rom3IconImage         *ebiten.Image
	scale1IconImage       *ebiten.Image
	scale2IconImage       *ebiten.Image
	keyboard              *ebiten.Image
	font                  font.Face
}

func LoadPNGAsset(fileName string) *ebiten.Image {
	file, err := assetsFS.Open(fileName)
	if err != nil {
		panic(err)
	}

	decoded, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(decoded)
}

func LoadTTFAsset(fileName string, size, dpi float64) font.Face {
	data, err := assetsFS.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	ttf, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}

	font, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		panic(err)
	}

	return font
}

func NewResources() Resources {
	return Resources{
		tapeIconImage:         LoadPNGAsset("assets/cassette.png"),
		keyboardUpIconImage:   LoadPNGAsset("assets/keyboard_up.png"),
		keyboardDownIconImage: LoadPNGAsset("assets/keyboard_down.png"),
		volumeIconImage:       LoadPNGAsset("assets/volume.png"),
		muteIconImage:         LoadPNGAsset("assets/mute.png"),
		cpu1IconImage:         LoadPNGAsset("assets/cpu1.png"),
		cpu2IconImage:         LoadPNGAsset("assets/cpu2.png"),
		cpu3IconImage:         LoadPNGAsset("assets/cpu3.png"),
		rom1IconImage:         LoadPNGAsset("assets/rom1.png"),
		rom2IconImage:         LoadPNGAsset("assets/rom2.png"),
		rom3IconImage:         LoadPNGAsset("assets/rom3.png"),
		scale1IconImage:       LoadPNGAsset("assets/scale1.png"),
		scale2IconImage:       LoadPNGAsset("assets/scale2.png"),
		keyboard:              LoadPNGAsset("assets/primo_zold.png"),
		font:                  LoadTTFAsset("assets/Roboto-Regular.ttf", 16, 72),
	}
}
