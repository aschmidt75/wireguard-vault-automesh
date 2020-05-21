VERSION := `cat VERSION`
SOURCES ?= $(shell find . -name "*.go" -type f)
BINARY_NAME = wgvam

all: clean vet lint build

.PHONY: build
build:
	go build -o release/${BINARY_NAME} -ldflags="-X main.version=${VERSION}" *.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o release/${BINARY_NAME}-linux-amd64 -ldflags="-X main.version=${VERSION}" *.go
vet:
	@go vet ./...

lint:
	@for file in ${SOURCES} ;  do \
		golint $$file ; \
	done

.PHONY: test
test:
	@go test -v ./...

.PHONY: cover
cover:
	@go test -coverprofile=cover.out ./...
	@go tool cover -func=cover.out

.PHONY: clean
clean:
	@rm -rf release/*
	@rm -f cover.out
