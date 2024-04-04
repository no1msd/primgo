package roms

import (
	_ "embed"
)

//go:embed a64.bin
var A64 []byte

//go:embed b64.bin
var B64 []byte

//go:embed c64.bin
var C64 []byte
