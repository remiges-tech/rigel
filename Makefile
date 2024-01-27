.PHONY: build-wsc-dev run-wsc-dev build-rigelctl

build-wsc-dev:
	mkdir -p out
	go build -tags dev -o out/rigel-server.exe server/main.go

run-wsc-dev: build-wsc-dev
	./out/rigel-server.exe

build-rigelctl:
	mkdir -p out
	go build -o out/rigelctl cmd/rigelctl/main.go