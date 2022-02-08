.PHONY: build clean tool lint help

all: build

build:
	GOOS=linux GOARCH=amd64 go build -o ./bin/dbcompare main.go
	
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