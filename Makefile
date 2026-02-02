# GoLemming Makefile

APP_NAME := golemming
BINARY_NAME := golemming.exe
BIN_DIR := bin

VERSION ?=
GITHUB_REPO := $(shell git remote get-url origin 2>/dev/null | sed 's/.*github.com[:/]\(.*\)\.git/\1/' || echo "thesimpledev/golemming")

.PHONY: help build build-linux debug clean release

help:
	@echo "GoLemming Build System"
	@echo ""
	@echo "Commands:"
	@echo "  make build                - Build for Windows (cross-compile)"
	@echo "  make build-linux          - Build for Linux"
	@echo "  make debug                - Quick Windows build with debug symbols"
	@echo "  make release VERSION=x.x.x - Build and release to GitHub"
	@echo "  make clean                - Remove build artifacts"
	@echo ""
	@echo "Example:"
	@echo "  make build"
	@echo "  make release VERSION=1.0.0"
	@echo ""
	@echo "Stable download URL (always points to latest release):"
	@echo "  https://github.com/$(GITHUB_REPO)/releases/latest/download/$(BINARY_NAME)"

build:
	@echo "=== Building $(APP_NAME) for Windows ==="
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/golemming
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)"

build-linux:
	@echo "=== Building $(APP_NAME) for Linux ==="
	mkdir -p $(BIN_DIR)
	go build -ldflags "-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/golemming
	@echo "Built: $(BIN_DIR)/$(APP_NAME)"

debug:
	@echo "=== Building $(APP_NAME) (debug) for Windows ==="
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/golemming
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)"

release:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=x.x.x)
endif
	@echo "=== Building $(APP_NAME) v$(VERSION) for Windows ==="
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/golemming
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "=== Creating GitHub Release v$(VERSION) ==="
	@if gh release view v$(VERSION) --repo $(GITHUB_REPO) >/dev/null 2>&1; then \
		echo "Release v$(VERSION) exists, updating..."; \
		gh release upload v$(VERSION) $(BIN_DIR)/$(BINARY_NAME) --repo $(GITHUB_REPO) --clobber; \
	else \
		echo "Creating release v$(VERSION)..."; \
		gh release create v$(VERSION) $(BIN_DIR)/$(BINARY_NAME) \
			--repo $(GITHUB_REPO) \
			--title "$(APP_NAME) v$(VERSION)" \
			--notes "## $(APP_NAME) v$(VERSION)"; \
	fi
	@echo ""
	@echo "=== Done! ==="
	@echo "Release: https://github.com/$(GITHUB_REPO)/releases/tag/v$(VERSION)"
	@echo "Latest:  https://github.com/$(GITHUB_REPO)/releases/latest/download/$(BINARY_NAME)"

clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	@echo "Done"
