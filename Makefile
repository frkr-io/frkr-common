.PHONY: build clean test

# Build the migrate binary to the bin folder
build:
	@mkdir -p bin
	go build -o bin/migrate ./cmd/migrate

# Clean the bin folder
clean:
	rm -rf bin

# Run tests
test:
	go test ./...

