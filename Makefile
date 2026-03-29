.PHONY: build run test clean

build:
	go build -o bin/clearpaste ./cmd/clearpaste

run:
	go run ./cmd/clearpaste

test:
	go test ./... -v

clean:
	rm -rf bin/
