BINARY = skill-chrome-host
VERSION = 0.1.0

.PHONY: build build-all clean test

build:
	go build -ldflags="-s -w" -o dist/$(BINARY) .

build-all:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY)-windows-amd64.exe .

test:
	go test -v ./...

clean:
	rm -rf dist/
