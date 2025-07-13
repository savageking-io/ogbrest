BINARY_NAME=ogbrest
BIN_DIR=./bin
PLATFORMS=linux darwin windows
ARCH=amd64
VERSION=$(shell git describe --tags 2>/dev/null || cat VERSION)
LDFLAGS=-ldflags "-X main.AppVersion=$(VERSION)"

.PHONY: all
all: clean build

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@mkdir -p $(BIN_DIR)

.PHONY: build
build: $(PLATFORMS)

.PHONY: linux
linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=$(ARCH) go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-$(ARCH) .

.PHONY: darwin
darwin:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=$(ARCH) go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-$(ARCH) .

.PHONY: windows
windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=$(ARCH) go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-windows-$(ARCH).exe .

.PHONY: version
version:
	@echo "Version: $(VERSION)"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all      - Clean and build for all platforms"
	@echo "  clean    - Remove the bin directory"
	@echo "  build    - Build for all platforms"
	@echo "  linux    - Build for Linux"
	@echo "  darwin   - Build for macOS"
	@echo "  windows  - Build for Windows"
	@echo "  version  - Show the current version"
	@echo "  help     - Show this help message"