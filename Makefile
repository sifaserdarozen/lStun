
.PHONY: build test fmt docker-build docker-run clean

VERSION:=$(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE:=$(shell date +%Y-%m-%dT%H:%M:%S)
DIRTY:=$(shell ! [ -z "`git status --porcelain=v1 2>/dev/null`" ] && echo "dirty-")
BUILD_VERSION:=${DIRTY}${VERSION}

LDFLAGS += -X "github.com/sifaserdarozen/stun/stun.Version=$(BUILD_VERSION)"
LDFLAGS += -X "github.com/sifaserdarozen/stun/stun.BuildDate=$(BUILD_DATE)"

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin ./...

test:
	go test ./...

fmt:
	go fmt ./... && go vet ./... && golangci-lint run

docker-build:
	docker build -t stun -f docker/Dockerfile .

docker-run: docker-build
	docker run -p 3478:3478/udp stun:latest ./stun

clean:
	rm -rf ./bin/*