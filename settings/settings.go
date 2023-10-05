//go:build !wasm

package settings

import (
	"fmt"
	"os"
	"path/filepath"
)

func storagePath() string {
	return filepath.Join(localConfigDir(), "primgo", "settings.json")
}

func Save(data string) error {
	path := storagePath()
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return fmt.Errorf("cannot create settings directories: %w", err)
	}
	err = os.WriteFile(path, []byte(data), 0600)
	if err != nil {
		return fmt.Errorf("cannot write settings file: %w", err)
	}
	return nil
}

func Load() (string, error) {
	data, err := os.ReadFile(storagePath())
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("cannot read settings file: %w", err)
	}
	return string(data), nil
}
