.PHONY: build run test clean docker-build docker-run help

# Default key for local development only - change in production!
STORAGE_KEY ?= my-super-secret-key-for-dev-32!

help:
	@echo "SecureStorage-Go Makefile Commands"
	@echo ""
	@echo "make build        - Build the Go binary"
	@echo "make run          - Run the server locally"
	@echo "make test         - Run tests"
	@echo "make clean        - Remove build artifacts"
	@echo "make docker-build - Build Docker image"
	@echo "make docker-run   - Run Docker container"
	@echo ""

build:
	@echo "Building server..."
	go build -o bin/server ./cmd/server

run: build
	@echo "Starting server..."
	STORAGE_KEY="$(STORAGE_KEY)" ./bin/server

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf data/

docker-build:
	@echo "Building Docker image..."
	docker build -t secure-storage-go .

docker-run:
	@echo "Running Docker container..."
	docker run -d \
		--name secure-storage \
		-p 8080:8080 \
		-e STORAGE_KEY="$(STORAGE_KEY)" \
		-v $(PWD)/data:/app/data \
		secure-storage-go

docker-stop:
	@echo "Stopping container..."
	docker stop secure-storage || true
	docker rm secure-storage || true

# Run a quick test upload
test-upload:
	@echo "Testing file upload..."
	echo "Hello, this is a test file!" > /tmp/testfile.txt
	curl -X POST -F "file=@/tmp/testfile.txt" http://localhost:8080/upload
	@echo ""

# Run a quick test download
test-download:
	@echo "Testing file download..."
	curl http://localhost:8080/download/testfile.txt --output /tmp/downloaded.txt
	cat /tmp/downloaded.txt
	@echo ""
