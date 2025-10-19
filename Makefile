VERSION := $(shell git describe --tags --always)
COMMIT  := $(shell git rev-parse --short HEAD)
LDFLAGS := -X main.build_version=$(VERSION) -X main.build_commit=$(COMMIT)

default:
	go build -ldflags "$(LDFLAGS)"

all: linux_amd64 linux_arm64 win_amd64 win_arm64 darwin_arm64


linux_amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o note_linux_amd64

linux_arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o note_linux_arm64

win_amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o note_win_amd64.exe

win_arm64:
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o note_win_arm64.exe

darwin_arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o note_darwin_arm64.exe
