#!/usr/bin/make -f

.ONESHELL:
.SHELL := /usr/bin/bash

AUTHOR := "noelruault"
PROJECTNAME := $(shell basename "$$(pwd)")
PROJECTPATH := $(shell pwd)

help:
	@echo "Usage: make [options] [arguments]\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

example: ## Run a example using the default config and template (templates/input.yml)
	@mkdir -p $(PROJECTPATH)/tmp
	@go run cmd/main.go --output=tmp/example.pdf

font: ## Converts a given ttf file, to a compatible font, using the iso-8859-15.map by default
# Executable is installed in the directory named by the GOBIN environment variable which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH environment variable is not set
	$(eval FONT_FILE=$(shell bash -c 'read -p "Enter font file name, that should be located at $(PROJECTPATH)/fonts/: " FONT_FILE; echo $$FONT_FILE'))
	$(eval GOPDF_VERSION=$(shell sh -c "go list -m -u github.com/jung-kurt/gofpdf" | awk '{print $$2}'))
	$(eval GOPATH=$(shell sh -c "go env GOPATH"))
	$(eval MAKEFONT_PATH="$(GOPATH)/pkg/mod/github.com/jung-kurt/gofpdf@$(GOPDF_VERSION)")
	@cd $(PROJECTPATH)/fonts && go run $(MAKEFONT_PATH)/makefont/makefont.go --embed -enc=$(MAKEFONT_PATH)/font/iso-8859-15.map $(PROJECTPATH)/fonts/$(FONT_FILE)

release: ## Tags to trigger a new release
	@read -p "Release version: " VERSION;\
	git tag $$VERSION && git push origin --tags
