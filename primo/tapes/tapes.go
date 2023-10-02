package tapes

import (
	"embed"
)

//go:embed *.ptp
var tapeFS embed.FS

func ByName(name string) []byte {
	data, err := tapeFS.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return data
}
