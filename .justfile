#! language=just

test:
    echo "Running tests..."
    go test  -json ./... -cover | tparse -all
