VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/chris-roerig/pixelize/internal/version.Version=$(VERSION)"

.PHONY: build test release

build:
	go build $(LDFLAGS) -o pixelize ./cmd/pixelize

test:
	go test ./...

# Create an annotated tag. Usage: make release VERSION=v0.2.0
release:
	@echo "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+' || \
		{ echo "VERSION must look like vX.Y.Z (got '$(VERSION)')"; exit 1; }
	git tag -a $(VERSION) -m "$(VERSION)"
	@echo "Created tag $(VERSION). Push with: git push origin main --follow-tags"
