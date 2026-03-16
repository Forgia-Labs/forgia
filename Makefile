VERSION ?= dev

.PHONY: build test lint clean build-linux

build:
	go build -ldflags "-X github.com/Deepzima/forgia/cmd/forgia/cmd.Version=$(VERSION)" -o build/forgia ./cmd/forgia

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf build/

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/Deepzima/forgia/cmd/forgia/cmd.Version=$(VERSION)" -o build/forgia-linux-amd64 ./cmd/forgia
