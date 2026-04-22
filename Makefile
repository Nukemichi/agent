APP_NAME=agent-michi
BUILD_DIR=build

.PHONY: all linux-amd64 linux-arm64 linux-armv7 darwin-arm64 clean

all: linux-amd64 linux-arm64 linux-armv7 darwin-arm64

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/agent

linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/agent

linux-armv7:
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 ./cmd/agent

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/agent

clean:
	rm -rf $(BUILD_DIR)
