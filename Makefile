VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
TARBALL := bin/mage-$(VERSION)-$(shell go env GOOS)_$(shell go env GOARCH).tar.gz
GOFLAGS := -ldflags='-s -w -X main.Version=$(VERSION)'
TARARGS := $$(test $$(uname -s) = Linux && echo "--owner=0 --group=0" || echo "--uid=0 --gid=0")

.PHONY: all

all: $(TARBALL)

clean:
	rm -f bin/*

$(TARBALL): bin/mage
	tar -C bin -zcf $(TARBALL) --numeric-owner $(TARARGS) ./mage

bin/%: *.go go.*
	mkdir -p bin
	go build -o $@ $(GOFLAGS)
