# Google Form Submitter Makefile
# 支援 CGO 跨平台編譯

# 變數定義
APP_NAME=google-form-submitter
VERSION?=1.0.0
BUILD_DIR=build
GO=go
GOFLAGS=-v

# 取得當前作業系統和架構
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# 輸出檔案名稱
BINARY_NAME=$(APP_NAME)
BINARY_DARWIN=$(BINARY_NAME)-darwin-$(GOARCH)
BINARY_WINDOWS=$(BINARY_NAME)-windows-$(GOARCH).exe

# 編譯標記
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

# 顏色輸出
COLOR_RESET=\033[0m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: all clean build build-darwin build-windows test run help deps

# 預設目標
all: clean build

# 顯示幫助訊息
help:
	@echo "$(COLOR_BLUE)Google Form Submitter - Makefile 指令$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)可用指令:$(COLOR_RESET)"
	@echo "  make build          - 編譯當前平台的執行檔"
	@echo "  make build-darwin   - 編譯 macOS 執行檔 (需要在 macOS 上執行)"
	@echo "  make build-windows  - 編譯 Windows 執行檔 (需要交叉編譯工具)"
	@echo "  make build-all      - 編譯所有平台的執行檔"
	@echo "  make run            - 編譯並執行程式"
	@echo "  make test           - 執行測試"
	@echo "  make clean          - 清理編譯產物"
	@echo "  make deps           - 安裝相依套件"
	@echo "  make help           - 顯示此幫助訊息"
	@echo ""

# 建立 build 目錄
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# 安裝相依套件
deps:
	@echo "$(COLOR_YELLOW)安裝相依套件...$(COLOR_RESET)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(COLOR_GREEN)相依套件安裝完成$(COLOR_RESET)"

# 編譯當前平台
build: $(BUILD_DIR) deps
	@echo "$(COLOR_YELLOW)編譯 $(GOOS)/$(GOARCH) 版本...$(COLOR_RESET)"
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(COLOR_GREEN)編譯完成: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

# 編譯 macOS 版本 (需要在 macOS 上執行)
build-darwin: $(BUILD_DIR) deps
	@echo "$(COLOR_YELLOW)編譯 macOS 版本...$(COLOR_RESET)"
ifeq ($(GOOS),darwin)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN)-amd64 .
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN)-arm64 .
	@echo "$(COLOR_GREEN)macOS 版本編譯完成$(COLOR_RESET)"
	@echo "  - $(BUILD_DIR)/$(BINARY_DARWIN)-amd64 (Intel)"
	@echo "  - $(BUILD_DIR)/$(BINARY_DARWIN)-arm64 (Apple Silicon)"
else
	@echo "$(COLOR_YELLOW)警告: macOS 版本需要在 macOS 系統上編譯$(COLOR_RESET)"
endif

# 編譯 Windows 版本 (需要 mingw-w64)
build-windows: $(BUILD_DIR) deps
	@echo "$(COLOR_YELLOW)編譯 Windows 版本...$(COLOR_RESET)"
ifeq ($(GOOS),darwin)
	@echo "$(COLOR_YELLOW)檢查 mingw-w64 是否已安裝...$(COLOR_RESET)"
	@which x86_64-w64-mingw32-gcc > /dev/null || (echo "$(COLOR_YELLOW)請先安裝 mingw-w64: brew install mingw-w64$(COLOR_RESET)" && exit 1)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) .
	@echo "$(COLOR_GREEN)Windows 版本編譯完成: $(BUILD_DIR)/$(BINARY_WINDOWS)$(COLOR_RESET)"
else ifeq ($(GOOS),windows)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) .
	@echo "$(COLOR_GREEN)Windows 版本編譯完成: $(BUILD_DIR)/$(BINARY_WINDOWS)$(COLOR_RESET)"
else
	@echo "$(COLOR_YELLOW)警告: Windows 交叉編譯需要 mingw-w64 工具鏈$(COLOR_RESET)"
endif

# 編譯所有平台
build-all: build-darwin build-windows
	@echo "$(COLOR_GREEN)所有平台編譯完成$(COLOR_RESET)"
	@ls -lh $(BUILD_DIR)/

# 執行測試
test:
	@echo "$(COLOR_YELLOW)執行測試...$(COLOR_RESET)"
	CGO_ENABLED=1 $(GO) test -v ./...
	@echo "$(COLOR_GREEN)測試完成$(COLOR_RESET)"

# 編譯並執行
run: build
	@echo "$(COLOR_YELLOW)啟動程式...$(COLOR_RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# 清理編譯產物
clean:
	@echo "$(COLOR_YELLOW)清理編譯產物...$(COLOR_RESET)"
	rm -rf $(BUILD_DIR)
	$(GO) clean
	@echo "$(COLOR_GREEN)清理完成$(COLOR_RESET)"

# 顯示版本資訊
version:
	@echo "$(COLOR_BLUE)版本: $(VERSION)$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Go 版本: $(shell $(GO) version)$(COLOR_RESET)"
