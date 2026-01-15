# 編譯指南

本專案使用 Makefile 來管理編譯流程，支援 macOS 和 Windows 平台的跨平台編譯。

## 前置需求

### macOS
```bash
# 安裝 Xcode Command Line Tools (提供 CGO 所需的 C 編譯器)
xcode-select --install

# 如需編譯 Windows 版本，需安裝 mingw-w64
brew install mingw-w64
```

### Windows
```bash
# 需要安裝 MinGW-w64 或 TDM-GCC
# 下載位置: https://www.mingw-w64.org/
# 或使用 MSYS2: https://www.msys2.org/
```

## 快速開始

### 查看所有可用指令
```bash
make help
```

### 編譯當前平台
```bash
make build
```
編譯完成後，執行檔位於 `build/google-form-submitter`

### 編譯並執行
```bash
make run
```

## 跨平台編譯

### 在 macOS 上編譯 macOS 版本
```bash
make build-darwin
```
會產生兩個版本：
- `build/google-form-submitter-darwin-amd64` (Intel Mac)
- `build/google-form-submitter-darwin-arm64` (Apple Silicon)

### 在 macOS 上編譯 Windows 版本
```bash
# 先安裝 mingw-w64
brew install mingw-w64

# 編譯 Windows 版本
make build-windows
```
會產生：`build/google-form-submitter-windows-amd64.exe`

### 編譯所有平台
```bash
make build-all
```

## 其他指令

### 安裝相依套件
```bash
make deps
```

### 執行測試
```bash
make test
```

### 清理編譯產物
```bash
make clean
```

## 注意事項

1. **CGO 必須啟用**: 本專案使用 SQLite (mattn/go-sqlite3)，需要 CGO 支援
2. **交叉編譯限制**: 
   - macOS 版本建議在 macOS 上編譯
   - Windows 版本可在 macOS 上透過 mingw-w64 交叉編譯
3. **編譯器需求**: 確保系統已安裝 C 編譯器 (gcc/clang)

## 疑難排解

### 錯誤: "gcc: command not found"
- macOS: 執行 `xcode-select --install`
- Windows: 安裝 MinGW-w64 或 TDM-GCC

### 錯誤: "x86_64-w64-mingw32-gcc: command not found"
在 macOS 上編譯 Windows 版本時需要：
```bash
brew install mingw-w64
```

### CGO 相關錯誤
確認環境變數：
```bash
# 檢查 CGO 是否啟用
go env CGO_ENABLED

# 應該顯示 "1"
```

## 部署

編譯完成後，需要將以下檔案一起部署：
- 執行檔 (從 `build/` 目錄)
- `config.json` (配置檔)
- `views/` 目錄 (HTML 模板)
- `data.db` 會自動建立

範例目錄結構：
```
deployment/
├── google-form-submitter (或 .exe)
├── config.json
├── views/
│   ├── index.html
│   └── result.html
└── data.db (執行後自動產生)
```
