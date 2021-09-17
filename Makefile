EXTLDFLAGS:=-w -extldflags "-static"

export PATH:=$(shell pwd)/bin:$(PATH)

# Mock source packages
GO_MOCK_SOURCES := ./cmd/...

path:=$(PATH)
export PATH:=./bin:$(path)
.PHONY: test
test: unit e2e

.PHONY: unit
unit:
	go test $(go list ./... | grep -v /test/)

.PHONY: e2e
e2e: build
	go test ./test/e2e -v

.PHONY: build
build: generate
	CGO_ENABLED=0 go build -a -tags netgo -ldflags "$(EXTLDFLAGS)" -o bin/pvsadm .

.PHONY: generate
generate:
	go generate ${GO_MOCK_SOURCES}
