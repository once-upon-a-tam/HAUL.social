set dotenv-load

# Run the app locally
run:
  go run main.go serve

# Run tests
test:
    go test ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Tidy dependencies
tidy:
    go mod tidy

# Run linters
lint:
    golangci-lint run ./...
