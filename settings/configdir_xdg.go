//go:build !windows && !darwin && !wasm

package settings

import (
	"os"
	"path/filepath"
)

func localConfigDir() string {
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return os.Getenv("XDG_CONFIG_HOME")
	}
	return filepath.Join(os.Getenv("HOME"), ".config")
}
