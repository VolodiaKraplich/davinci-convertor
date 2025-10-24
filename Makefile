# CrystalNetwork Studio
# <------------------->
# davinci-convert Makefile
# Performance-optimized
# <------------------->


PROJECT_NAME := davinci-convert
BINARY := $(PROJECT_NAME)
BIN_DIR := bin
GO := go
GOCACHE := $(shell $(GO) env GOCACHE)
export GOCACHE

BUILD_FLAGS := -trimpath
LDFLAGS := -ldflags="-s -w -buildid="

GO_SOURCES := $(shell find . -path ./vendor -prune -o -name '*.go' -print)

.DEFAULT_GOAL := build

# Main build target: build and compress
build: | $(BIN_DIR)
	@echo "Building optimized $(PROJECT_NAME)..."
	CGO_ENABLED=0 $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) ./main.go
	@echo "Optimized build completed."

	@echo "Checking for UPX..."
	@which upx > /dev/null || (echo "Error: UPX not found. Please install UPX." && exit 1)
	@echo "Compressing binary with UPX..."
	upx --best --lzma $(BIN_DIR)/$(BINARY)
	@echo "UPX compression done: $(BIN_DIR)/$(BINARY)"

$(BIN_DIR):
	mkdir -p $@

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -f $(BINARY).{bash,zsh,fish}
	$(GO) clean -cache -modcache
	@echo "Cleanup completed"

fmt:
	@echo "Formatting Go code..."
	$(GO) fmt ./...
	@echo "Code formatted"

completions: build
	@echo "Generating shell completions..."
	$(BIN_DIR)/$(BINARY) completion bash > $(BINARY).bash
	$(BIN_DIR)/$(BINARY) completion zsh > $(BINARY).zsh
	$(BIN_DIR)/$(BINARY) completion fish > $(BINARY).fish
	@echo "Shell completions generated: $(BINARY).{bash,zsh,fish}"

install-completions: completions
	@echo "Installing shell completions..."
	@# Bash completion
	@if [ -d /usr/share/bash-completion/completions ]; then \
		install -Dm644 $(BINARY).bash /usr/share/bash-completion/completions/$(BINARY); \
		echo "Bash completion installed to /usr/share/bash-completion/completions/"; \
	elif [ -d ~/.local/share/bash-completion/completions ]; then \
		install -Dm644 $(BINARY).bash ~/.local/share/bash-completion/completions/$(BINARY); \
		echo "Bash completion installed to ~/.local/share/bash-completion/completions/"; \
	else \
		mkdir -p ~/.local/share/bash-completion/completions; \
		install -Dm644 $(BINARY).bash ~/.local/share/bash-completion/completions/$(BINARY); \
		echo "Bash completion installed to ~/.local/share/bash-completion/completions/"; \
	fi
	@# Zsh completion
	@if [ -d /usr/share/zsh/site-functions ]; then \
		install -Dm644 $(BINARY).zsh /usr/share/zsh/site-functions/_$(BINARY); \
		echo "Zsh completion installed to /usr/share/zsh/site-functions/"; \
	else \
		mkdir -p ~/.local/share/zsh/site-functions; \
		install -Dm644 $(BINARY).zsh ~/.local/share/zsh/site-functions/_$(BINARY); \
		echo "Zsh completion installed to ~/.local/share/zsh/site-functions/"; \
		echo "Add this to your .zshrc if not already present:"; \
		echo "  fpath=(~/.local/share/zsh/site-functions \$$fpath)"; \
	fi
	@# Fish completion
	@mkdir -p ~/.config/fish/completions
	@install -Dm644 $(BINARY).fish ~/.config/fish/completions/$(BINARY).fish
	@echo "Fish completion installed to ~/.config/fish/completions/"
	@echo ""
	@echo "Shell completions installed successfully!"
	@echo "You may need to restart your shell or source the completion files."

help:
	@echo ""
	@echo "$(PROJECT_NAME) Makefile"
	@echo "Usage: make [target]"
	@echo ""
	@echo "  build                 Build and compress binary with UPX"
	@echo "  clean                 Clean build artifacts and completions"
	@echo "  fmt                   Format Go code"
	@echo "  completions           Generate shell completions (bash, zsh, fish)"
	@echo "  install-completions   Install shell completions to system/user directories"
	@echo "  help                  Show this help"
	@echo ""

.PHONY: build clean fmt completions install-completions help
