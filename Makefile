BINARY_NAME := grpc-pubsub
MAIN_PACKAGE ?= ./cmd/pubsub/main.go

BUILD_DIR := dist
EXTRA_BUILD_FLAGS ?=
BUILD_FLAGS := -v -ldflags="-s -w"

PROTO_DIR := pkg/protobuf
PROTO_OUT_DIR := internal/generated/__proto__
MODULE_PATH := github.com/iamBelugaa/grpc-pubsub

# ANSI Color Codes
GREEN := \033[32m
YELLOW := \033[33m
CYAN := \033[36m
RESET := \033[0m

all: build run

build: tidy gen-pb
	@echo "$(CYAN) Building $(BINARY_NAME) for $(shell go env GOOS)/$(shell go env GOARCH)...$(RESET)"
	@GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build $(BUILD_FLAGS) $(EXTRA_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN) Build complete.$(RESET)"

run: build
	@echo "$(CYAN) Running $(BINARY_NAME) $(RUN_ARGS)...$(RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) $(RUN_ARGS)

tidy:
	@echo "$(CYAN) Tidying Go modules...$(RESET)"
	@go mod tidy
	@echo "$(GREEN) Go modules tidied.$(RESET)"

deps:
	@echo "$(CYAN) Downloading Go modules...$(RESET)"
	@go mod download
	@go mod verify
	@echo "$(GREEN) Go modules downloaded.$(RESET)"

fmt:
	@echo "$(CYAN) Formatting Go code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN) Formatting complete.$(RESET)"

clean:
	@echo "$(YELLOW) Cleaning build artifacts...$(RESET)"
	@go clean
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN) Clean complete.$(RESET)"

gen-pb: clean-proto-gen
	@echo "$(CYAN) Generating Protocol Buffer and GRPC Go code...$(RESET)"
	@mkdir -p $(PROTO_OUT_DIR)
	@protoc \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=module=$(MODULE_PATH) \
		--go-grpc_out=$(PROTO_OUT_DIR) \
		--proto_path=$(PROTO_DIR) \
		--go-grpc_opt=module=$(MODULE_PATH) \
		$(PROTO_DIR)/pubsub.proto
	@echo "$(GREEN) Protocol Buffer and GRPC generation complete$(RESET)"

clean-proto-gen:
	@echo "$(YELLOW) Cleaning previous Protocol Buffer and GRPC generated files...$(RESET)"
	@rm -rf $(PROTO_OUT_DIR)
	@echo "$(GREEN) Cleanup complete$(RESET)"

