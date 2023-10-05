package dialog

import (
	"syscall/js"
)

func BrowseFile() chan *OpenedFile {
	res := make(chan *OpenedFile)

	fileInput := js.Global().Get("document").Call("createElement", "input")
	fileInput.Set("type", "file")

	fileInput.Call("addEventListener", "change", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		files := fileInput.Get("files")
		if files.Length() > 0 {
			selectedFile := files.Index(0)
			fileName := selectedFile.Get("name").String()
			reader := js.Global().Get("FileReader").New()
			reader.Call("addEventListener", "load", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
				fileContent := js.Global().Get("Uint8Array").New(reader.Get("result"))
				data := make([]byte, fileContent.Length())
				js.CopyBytesToGo(data, fileContent)
				go func() {
					res <- &OpenedFile{Data: data, Name: fileName}
				}()
				return nil
			}))
			reader.Call("addEventListener", "error", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
				go func() {
					res <- nil
				}()
				return nil
			}))
			reader.Call("readAsArrayBuffer", selectedFile)
		} else {
			go func() {
				res <- nil
			}()
		}
		return nil
	}))

	fileInput.Call("click")

	return res
}
