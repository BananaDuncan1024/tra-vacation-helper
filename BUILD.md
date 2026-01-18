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

## GitHub Releases 下載

如果不想自行編譯，可以直接從 GitHub Releases 頁面下載預編譯版本。

### 下載步驟

1. 前往專案的 [Releases](../../releases) 頁面
2. 選擇最新版本
3. 根據你的作業系統下載對應的 ZIP 檔案：
   - `google-form-submitter-darwin-amd64.zip` - macOS Intel
   - `google-form-submitter-darwin-arm64.zip` - macOS Apple Silicon
   - `google-form-submitter-windows-amd64.zip` - Windows 64-bit

4. 解壓縮後，將 `config.json.example` 複製為 `config.json` 並修改設定

### macOS 執行權限

在 macOS 上下載後，需要賦予執行權限：
```bash
chmod +x google-form-submitter
```

如果遇到「無法打開，因為無法驗證開發者」的提示，可以：
1. 右鍵點擊執行檔 → 打開
2. 或在終端機執行：`xattr -d com.apple.quarantine google-form-submitter`

## 發布新版本

專案使用 GitHub Actions 自動化發布流程。

### 發布步驟

```bash
# 1. 確保所有更改已提交
git add .
git commit -m "準備發布 v1.0.0"

# 2. 建立版本標籤
git tag -a v1.0.0 -m "Release version 1.0.0"

# 3. 推送標籤到 GitHub
git push origin v1.0.0
```

推送標籤後，GitHub Actions 會自動：
1. 在 macOS 和 Windows 環境編譯程式
2. 建立包含執行檔、設定範例和模板的 ZIP 檔案
3. 建立 GitHub Release 並上傳所有檔案

### 版本命名規則

使用語義化版本號 (Semantic Versioning)：
- `v1.0.0` - 正式版本
- `v1.0.0-beta` - 測試版本
- `v1.0.0-rc1` - 候選版本
