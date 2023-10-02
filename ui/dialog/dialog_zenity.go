//go:build !wasm

package dialog

import (
	"os"
	"path/filepath"

	"github.com/ncruces/zenity"
)

func BrowseFile() chan *OpenedFile {
	res := make(chan *OpenedFile)

	go func() {
		fileName, err := zenity.SelectFile(
			zenity.FileFilters{
				{Name: "Primo tape files", Patterns: []string{"*.ptp"}, CaseFold: false},
			})
		if err != nil {
			res <- nil
			return
		}

		fileContent, err := os.ReadFile(fileName)
		if err != nil {
			res <- nil
			return
		}

		res <- &OpenedFile{Data: fileContent, Name: filepath.Base(fileName)}
	}()

	return res
}
