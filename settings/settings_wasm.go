package settings

import (
	"syscall/js"
)

const key = "primgodata"

func Save(data string) error {
	js.Global().Get("localStorage").Call("setItem", key, data)
	return nil
}

func Load() (string, error) {
	return js.Global().Get("localStorage").Call("getItem", key).String(), nil
}
