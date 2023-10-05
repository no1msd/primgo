WEB_DIR := dist-web

.PHONY: primgo web serve lint clean

primgo:
ifeq ($(shell go env GOOS),windows)
	go run github.com/tc-hib/go-winres@latest make
	go build -ldflags "-s -w -H windowsgui"
else
	go build -ldflags "-s -w"
endif

web:
	mkdir -p ${WEB_DIR}
	GOOS=js GOARCH=wasm go build -o ${WEB_DIR}/primgo.wasm
	cp $(shell go env GOROOT)/misc/wasm/wasm_exec.js ${WEB_DIR}
	cp web/* ${WEB_DIR}

serve:
	go run github.com/hajimehoshi/wasmserve@latest -http :8082 .

lint:
	golangci-lint run

clean:
	go clean
	rm -rf ${WEB_DIR}
	rm -f rsrc_windows_*.syso

default: primgo
