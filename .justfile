#! language=just

dev:
    @kill -9 $(lsof -ti:8080) || true
    go run cmd/main.go

test:
    echo "Running tests..."
    go test  -json ./... -cover | tparse -all
