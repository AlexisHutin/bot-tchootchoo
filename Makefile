# Variables
APP_NAME := tchootchoo
BUILD_DIR := build
BIN_PATH := $(BUILD_DIR)/$(APP_NAME)
REMOTE_USER := username               # Nom d'utilisateur SSH sur la machine distante
REMOTE_HOST := 192.168.x.x           # Adresse IP de la machine distante
REMOTE_DIR := /path/to/deploy/dir    # Répertoire sur la machine distante où le binaire sera copié
SSH_KEY := ~/.ssh/id_rsa             # Clé privée SSH à utiliser

# Default target
all: build upload

# Build the Go binary
build:
	@echo "Building the Go binary..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_PATH) .
	@echo "Build complete: $(BIN_PATH)"

# Upload the binary to the remote machine
upload: build
	@echo "Uploading the binary to $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)..."
	scp -i $(SSH_KEY) $(BIN_PATH) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_DIR)
	@echo "Upload complete."

# Clean the build directory
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Run the binary on the remote server
remote-run: upload
	@echo "Running the binary on the remote server..."
	ssh -i $(SSH_KEY) $(REMOTE_USER)@$(REMOTE_HOST) "cd $(REMOTE_DIR) && ./$(APP_NAME)"

# Full pipeline: Build, upload, and run
deploy: build upload remote-run
