.PHONY: build test lint clean

build:
	go build -o bin/contextual ./cmd/contextual/

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
