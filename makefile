# Makefile
.PHONY: build run test clean

build:
	go build -o letta ./cmd/letta

run:
	go run ./cmd/letta

test:
	go test ./...

clean:
	rm -f letta