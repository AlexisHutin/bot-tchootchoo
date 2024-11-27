include .env
export $(shell sed 's/=.*//' .env)

# Variables
BINARY_NAME=tchootchoo
BUILD_DIR=./build
CONFIG_FILE=config.yml
CRONTAB_FILE=bot.cron
ENV_FILE=.env
BIN_PATH=$(BUILD_DIR)/$(BINARY_NAME)

# Default target
all: build upload

.PHONY: build upload clean deploy

# Build the Go binary
build:
	@echo "Building the Go binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_PATH) .
	@echo "Build complete: $(BIN_PATH)"

# Upload the binary to the remote machine
upload:
	@echo "Uploading the binary to the remote server..."
	scp -i $(SSH_KEY) $(BUILD_DIR)/$(BINARY_NAME) $(SSH_USER)@$(SSH_HOST):$(REMOTE_DIR)
	scp -i $(SSH_KEY) $(CONFIG_FILE) $(SSH_USER)@$(SSH_HOST):$(REMOTE_DIR)
	scp -i $(SSH_KEY) $(ENV_FILE) $(SSH_USER)@$(SSH_HOST):$(REMOTE_DIR)
	@echo "Upload complete."

	@echo "Uploading crontab file to the remote server..."
	scp -i $(SSH_KEY) $(CRONTAB_FILE) $(SSH_USER)@$(SSH_HOST):$(REMOTE_DIR)/$(CRONTAB_FILE)
	@echo "Adding crontab to the server..."
	ssh -i $(SSH_KEY) $(SSH_USER)@$(SSH_HOST) "crontab $(REMOTE_DIR)/$(CRONTAB_FILE)"
	@echo "Crontab added successfully!"

# Clean the build directory
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Full pipeline: Build, upload, and run
deploy: build upload
