NAME      = ctop
VERSION  := $(shell cat VERSION)
BUILD    := $(shell git rev-parse --short HEAD)
LD_FLAGS  = "-w -X main.version=$(VERSION) -X main.build=$(BUILD)"
# sha256sum is absent on macOS; shasum -a 256 produces the same output format
SHA256   := $(shell command -v sha256sum 2>/dev/null || echo "shasum -a 256")

## all: run checks then build the local binary
all: check build

## fmt: format all Go sources
fmt:
	gofmt -w .

## vet: run static analysis
vet:
	go vet ./...

## test: run all tests with race detection
test:
	go test -race ./...

## check: vet + test, the gate for any commit
check: vet test

## build: build a local release binary
build:
	go mod download
	CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o $(NAME)

## build-all: cross-compile release binaries into _build/
build-all:
	mkdir -p _build
	GOOS=darwin  GOARCH=amd64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/ctop-$(VERSION)-darwin-amd64
	GOOS=linux   GOARCH=amd64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/ctop-$(VERSION)-linux-amd64
	GOOS=linux   GOARCH=arm     CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/ctop-$(VERSION)-linux-arm
	GOOS=linux   GOARCH=arm64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/ctop-$(VERSION)-linux-arm64
	GOOS=linux   GOARCH=ppc64le CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/ctop-$(VERSION)-linux-ppc64le
	GOOS=windows GOARCH=amd64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/ctop-$(VERSION)-windows-amd64
	cd _build && $(SHA256) * > sha256sums.txt

## run-dev: build without release tag and run with debug logging
run-dev:
	rm -f ctop.sock ctop
	go build -ldflags $(LD_FLAGS) -o ctop
	CTOP_DEBUG=1 ./ctop

## image: build the docker image
image:
	docker build -t ctop -f Dockerfile .

## release: verify build-all checksums and create a draft GitHub release
release:
	mkdir -p release
	cp _build/* release/
	cd release && $(SHA256) -c sha256sums.txt && \
	gh release create v$(VERSION) -d -t v$(VERSION) *

## clean: remove build artifacts
clean:
	rm -rf _build/ release/ ctop ctop.sock

.PHONY: all fmt vet test check build build-all run-dev image release clean
