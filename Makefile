.DEFAULT_GOAL := build
GOBIN ?= $(shell go env GOPATH)/bin

ifeq ($(filter-out /,$(abspath $(GOBIN))),)
$(error GOBIN is '$(GOBIN)'; it must name a real directory)
endif

.PHONY: build test clean push

build:
	go install ./...

test:
	go test ./... -count=1
	go vet ./...

clean:
	go clean ./...
	rm -f "$(GOBIN)"/*

push: build
	push vendor "$(GOBIN)/scarlettctl" scarlettctl
