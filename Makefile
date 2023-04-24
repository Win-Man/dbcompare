.PHONY: build clean tool lint help

REPO    := github.com/Win-Man/dbcompare

GOOS    := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH  := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOENV   := GO111MODULE=on CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO      := $(GOENV) go
GOBUILD := $(GO) build
GORUN   := $(GO) run
SHELL   := /usr/bin/env bash

COMMIT  := $(shell git describe --always --no-match --tags --dirty="-dev")
BUILDTS := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GITHASH := $(shell git rev-parse HEAD)
GITREF  := $(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS := -w -s
LDFLAGS += -X "$(REPO)/service.Version=$(COMMIT)"
LDFLAGS += -X "$(REPO)/service.BuildTS=$(BUILDTS)"
LDFLAGS += -X "$(REPO)/service.GitHash=$(GITHASH)"
LDFLAGS += -X "$(REPO)/service.GitBranch=$(GITREF)"

all: build

build:
	$(GOBUILD) -ldflags '$(LDFLAGS)'  -o ./bin/dbcompare_${GOARCH} main.go
	cp ./config/ctl.tmpl ./bin
	cp ./config/lightning_toml.tmpl ./bin
	cp ./config/sync-diff-config.tmpl ./bin
	
tool:
	go tool vet . |& grep -v vendor; true
	gofmt -w .

lint:
	golint ./...

clean:
	go clean -i ./...

help:
	@echo "make: compile packages and dependencies"
	@echo "make tool: run specified go tool"
	@echo "make lint: golint ./..."
	@echo "make clean: remove object files and cached files"