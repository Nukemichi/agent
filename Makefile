APP_NAME=agent-michi
BUILD_DIR=build

.PHONY: all linux-amd64 linux-arm64 clean

all: linux-amd64 linux-arm64

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/agent

linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/agent

clean:
	rm -rf $(BUILD_DIR)
