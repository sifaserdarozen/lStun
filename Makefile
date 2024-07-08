.PHONY: build test fmt docker-build docker-run clean

build:
	mkdir -p bin
	go build -o bin ./...

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