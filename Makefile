# Copyright 2019 The GoRE.tk Authors. All rights reserved.
# Use of this source code is governed by the license that
# can be found in the LICENSE file.

APP = redress

SHELL = /bin/bash
DIR = $(shell pwd)
GO = go

VERSION=$(shell git describe --tags 2> /dev/null || git rev-list -1 HEAD)
GOREVER=$(shell grep "goretk\/gore" go.mod | awk '{print $$2;}')
GOVER=$(shell go version | awk '{print $$3;}')
LDEXTRA=-X "main.redressVersion=$(VERSION)" -X "main.goreVersion=$(GOREVER)" -X "main.compilerVersion=$(GOVER)"

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m
MAKE_COLOR=\033[33;01m%-20s\033[0m

ARCH=GOARCH=amd64
TARGET_FOLDER=dist
PACKAGE=$(APP)-$(VERSION)
TAR_ARGS=cfz
RELEASE_FILES=LICENSE README.md
BUILD_OPTS=-ldflags='-s -w $(LDEXTRA)' -trimpath

# Linux options

LINUX_BUILD_FOLDER=build/linux
LINUX_ARCHIVE=$(TARGET_FOLDER)/$(PACKAGE)-linux-amd64.tar.gz
LINUX_BUILD_ART=$(LINUX_BUILD_FOLDER)/$(PACKAGE)/$(APP)
LINUX_GO_ENV=GOOS=linux $(ARCH)

# Windows options

WINDOWS_BUILD_FOLDER=build/windows
WINDOWS_ARCHIVE=$(APP)-$(VERSION)-windows.zip
WINDOWS_BUILD_ART=$(WINDOWS_BUILD_FOLDER)/$(PACKAGE)/$(APP).exe
WINDOWS_GO_ENV=GOOS=windows $(ARCH)

# macOS options

MACOS_BUILD_FOLDER=build/macos
MACOS_ARCHIVE=$(TARGET_FOLDER)/$(PACKAGE)-macos.tar.gz
MACOS_BUILD_ART=$(MACOS_BUILD_FOLDER)/$(PACKAGE)/$(APP)
MACOS_GO_ENV=GOOS=darwin $(ARCH)

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo -e "$(OK_COLOR)==== $(APP) [$(VERSION)] ====$(NO_COLOR)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(MAKE_COLOR) : %s\n", $$1, $$2}'

.PHONY: windows
windows: ## Make binary for Windows
	@echo -e "$(OK_COLOR)[$(APP)] Build for Windows$(NO_COLOR)"
	@$(WINDOWS_GO_ENV) $(GO) build -o $(APP).exe $(BUILD_OPTS) .

.PHONY: linux
linux: ## Make binary for Linux
	@echo -e "$(OK_COLOR)[$(APP)] Build for Linux$(NO_COLOR)"
	@$(LINUX_GO_ENV) $(GO) build -o $(APP) $(BUILD_OPTS) .

.PHONY: macos
macos: ## Make binary for macOS
	@echo -e "$(OK_COLOR)[$(APP)] Build for macOS$(NO_COLOR)"
	@$(MACOS_GO_ENV) $(GO) build -o $(APP) $(BUILD_OPTS) .

.PHONY: build
build: ## Make binary
	@echo -e "$(OK_COLOR)[$(APP)] Build$(NO_COLOR)"
	@$(GO) build -o $(APP) $(BUILD_OPTS) .

.PHONY: clean
clean: ## Remove build artifacts
	@echo -e "$(OK_COLOR)[$(APP)] Clean$(NO_COLOR)"
	@rm -rf build dist 2> /dev/null

.PHONY: release

$(LINUX_ARCHIVE): $(LINUX_BUILD_ART)
	@mkdir -p $(TARGET_FOLDER)
	@cp $(RELEASE_FILES) $(LINUX_BUILD_FOLDER)/$(PACKAGE)/.
	@tar $(TAR_ARGS) $(LINUX_ARCHIVE) -C $(LINUX_BUILD_FOLDER) $(PACKAGE)

$(MACOS_ARCHIVE): $(MACOS_BUILD_ART)
	@mkdir -p $(TARGET_FOLDER)
	@cp $(RELEASE_FILES) $(MACOS_BUILD_FOLDER)/$(PACKAGE)/.
	@tar $(TAR_ARGS) $(MACOS_ARCHIVE) -C $(MACOS_BUILD_FOLDER) $(PACKAGE)

$(WINDOWS_ARCHIVE): $(WINDOWS_BUILD_ART)
	@mkdir -p $(TARGET_FOLDER)
	@cp $(RELEASE_FILES) $(WINDOWS_BUILD_FOLDER)/$(PACKAGE)/.
	@cd $(WINDOWS_BUILD_FOLDER) && zip -r $(DIR)/$(TARGET_FOLDER)/$(WINDOWS_ARCHIVE) $(PACKAGE) > /dev/null

$(LINUX_BUILD_ART):
	@mkdir -p $(LINUX_BUILD_FOLDER)/$(PACKAGE)
	@echo -e "$(OK_COLOR)[$(APP)] Build for Linux$(NO_COLOR)"
	@$(LINUX_GO_ENV) $(GO) build -o $(LINUX_BUILD_ART) -v $(BUILD_OPTS) .

$(MACOS_BUILD_ART):
	@mkdir -p $(MACOS_BUILD_FOLDER)/$(PACKAGE)
	@echo -e "$(OK_COLOR)[$(APP)] Build for macOS$(NO_COLOR)"
	@$(MACOS_GO_ENV) $(GO) build -o $(MACOS_BUILD_ART) -v $(BUILD_OPTS) .

$(WINDOWS_BUILD_ART):
	@mkdir -p $(WINDOWS_BUILD_FOLDER)/$(PACKAGE)
	@echo -e "$(OK_COLOR)[$(APP)] Build for Windows$(NO_COLOR)"
	@$(WINDOWS_GO_ENV) $(GO) build -o $(WINDOWS_BUILD_ART) -v $(BUILD_OPTS) .

release: $(LINUX_ARCHIVE) $(WINDOWS_ARCHIVE) $(MACOS_ARCHIVE) ## Make release archives

