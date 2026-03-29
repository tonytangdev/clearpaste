.PHONY: build run test clean

build:
	mkdir -p cmd/clearpaste/icons
	cp icons/*.png cmd/clearpaste/icons/
	go build -o bin/clearpaste ./cmd/clearpaste

run:
	mkdir -p cmd/clearpaste/icons
	cp icons/*.png cmd/clearpaste/icons/
	go run ./cmd/clearpaste

test:
	go test ./... -v

clean:
	rm -rf bin/ cmd/clearpaste/icons/
