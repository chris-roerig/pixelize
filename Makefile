VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/chris-roerig/pixelize/internal/version.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o pixelize ./cmd/pixelize
