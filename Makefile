WEB_DIR := dist-web

.PHONY: primgo win web serve lint clean

primgo:
ifeq ($(shell go env GOOS),windows)
	go run github.com/tc-hib/go-winres@latest make
	go build -ldflags "-s -w -H windowsgui"
else
	go build -ldflags "-s -w"
endif

win:
	go run github.com/tc-hib/go-winres@latest make
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -H windowsgui"

web:
	mkdir -p ${WEB_DIR}
	GOOS=js GOARCH=wasm go build -o ${WEB_DIR}/primgo.wasm
	cp $(shell go env GOROOT)/misc/wasm/wasm_exec.js ${WEB_DIR}
	cp -r web/. ${WEB_DIR}
	cp ui/assets/icon512.png ${WEB_DIR}

serve:
	go run github.com/hajimehoshi/wasmserve@latest -http :8082 .

lint:
	golangci-lint run

clean:
	go clean
	rm -rf ${WEB_DIR}
	rm -f rsrc_windows_*.syso

default: primgo
