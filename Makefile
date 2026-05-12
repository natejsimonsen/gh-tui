BIN := gh-tui
BUILD_DIR := ./cmd/gh-tui
MOCK_DIR := testdata/mock

export GOTOOLCHAIN := auto

.PHONY: build test lint gen-mock run-mock bench clean

build:
	go build -o $(BIN) $(BUILD_DIR)

test:
	go test ./...

lint:
	golangci-lint run ./...

gen-mock:
	go run ./cmd/gen-mock-data -out $(MOCK_DIR) -prs 50

run-mock: gen-mock build
	./$(BIN) --mock-data $(MOCK_DIR)

bench: gen-mock
	go run ./cmd/bench -mock-data $(MOCK_DIR) -n 10

clean:
	rm -f $(BIN)
	rm -rf $(MOCK_DIR)
