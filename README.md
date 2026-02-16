# Google Form Submitter

[![CI](https://github.com/BananaDuncan1024/tra-vacation-helper/actions/workflows/ci.yml/badge.svg)](https://github.com/BananaDuncan1024/tra-vacation-helper/actions/workflows/ci.yml)
[![Release](https://github.com/BananaDuncan1024/tra-vacation-helper/actions/workflows/release.yml/badge.svg)](https://github.com/BananaDuncan1024/tra-vacation-helper/releases)

一個本地運行的請假申請系統，可自動提交表單到 Google Forms。支援網頁介面操作、API 呼叫，以及排程自動提交功能。

## ✨ 功能特色

- 🖥️ **網頁介面** - 簡潔的表單填寫頁面，支援立即提交與保存資料
- 📅 **排程管理** - 網頁 UI 即時管理排程，設定定時自動提交
- 🔄 **API 支援** - RESTful API 可整合其他系統
- 💾 **資料儲存** - SQLite 本地儲存，可預先保存表單資料供排程使用
- ⏰ **精準排程** - 指定日期時間自動提交，適合搶票/搶假場景
- 🔁 **自動重試** - 失敗時自動重試機制
- 🖥️ **跨平台** - 支援 macOS (Intel/Apple Silicon) 和 Windows

## 📥 下載

從 [Releases](https://github.com/BananaDuncan1024/tra-vacation-helper/releases) 頁面下載對應系統的預編譯版本：

| 平台 | 檔案 |
|------|------|
| macOS Intel | `google-form-submitter-darwin-amd64.zip` |
| macOS Apple Silicon | `google-form-submitter-darwin-arm64.zip` |
| Windows 64-bit | `google-form-submitter-windows-amd64.zip` |

## 🚀 快速開始

### 1. 下載並解壓縮

下載對應平台的 ZIP 檔案並解壓縮。

### 2. 設定 config.json

編輯 `config.json`，填入你的 Google Form 資訊：

```json
{
  "port": "8080",
  "form_url": "https://docs.google.com/forms/d/e/YOUR_FORM_ID/formResponse",
  "entry_map": {
    "name": "entry.123456789",
    "employee_id": "entry.987654321",
    "start_date": "entry.111111111",
    "end_date": "entry.222222222",
    "leave_type": "entry.333333333",
    "password": "entry.444444444"
  },
  "db_path": "data.db",
  "schedule": {
    "enabled": false,
    "date": "",
    "saved_form_id": 0,
    "prepare_seconds": 5,
    "retry_count": 3,
    "retry_interval": 100
  }
}
```

#### 如何取得 Google Form Entry ID

1. 開啟 Google Form 的填寫頁面
2. 按 `F12` 開啟開發者工具
3. 查看 HTML 原始碼，找到每個欄位的 `name` 屬性（格式為 `entry.XXXXXXXXX`）

### 3. 執行程式

**macOS:**
```bash
chmod +x google-form-submitter
./google-form-submitter
```

**Windows:**
```cmd
google-form-submitter.exe
```

### 4. 開啟瀏覽器

存取 `http://localhost:8080` 即可使用網頁介面。

## 🖥️ 網頁介面

系統提供兩個主要頁面，透過頂部導航列切換：

### 請假申請（`/`）

填寫請假表單後，可選擇兩種操作：

- **🚀 立即提交** - 直接提交到 Google Form
- **💾 保存資料** - 儲存至本地資料庫，可在排程管理中選用

### 排程管理（`/schedule`）

- 查看目前排程狀態（是否啟用、目標時間、重試設定）
- 選擇已保存的表單資料
- 設定排程日期，系統將在該日 00:00:00 自動提交
- 可隨時啟動或停止排程

### 操作流程

```
填寫表單 → 保存資料 → 排程管理選擇資料 → 設定日期 → 到期自動提交
```

## 📡 API 文件

### 提交表單

```http
POST /api/submit
Content-Type: application/json

{
  "name": "王小明",
  "employee_id": "A12345",
  "start_date": "2025-01-20",
  "end_date": "2025-01-22",
  "leave_type": "近假",
  "password": "your_password"
}
```

### 儲存表單資料

```http
POST /api/saved
Content-Type: application/json

{
  "name": "王小明",
  "employee_id": "A12345",
  "start_date": "2025-01-20",
  "end_date": "2025-01-22",
  "leave_type": "近假",
  "password": "your_password"
}
```

### 列出已儲存的表單

```http
GET /api/saved
```

### 取得單筆儲存的表單

```http
GET /api/saved/:id
```

### 刪除已儲存的表單

```http
DELETE /api/saved/:id
```

### 排程管理 API

| 方法 | 路徑 | 說明 |
|------|------|------|
| `GET` | `/api/schedule` | 取得排程狀態 |
| `POST` | `/api/schedule` | 建立並啟動排程 |
| `DELETE` | `/api/schedule` | 停止排程 |

#### 建立排程

```http
POST /api/schedule
Content-Type: application/json

{
  "date": "2025-01-20",
  "saved_form_id": 1,
  "prepare_seconds": 5,
  "retry_count": 3,
  "retry_interval": 100
}
```

| 參數 | 說明 |
|------|------|
| `date` | 排程日期（到達 00:00:00 時自動提交） |
| `saved_form_id` | 使用的儲存資料 ID |
| `prepare_seconds` | 提前準備秒數（預設 5） |
| `retry_count` | 失敗重試次數（預設 3） |
| `retry_interval` | 重試間隔，毫秒（預設 100） |

## 🔧 從原始碼編譯

請參閱 [BUILD.md](BUILD.md) 了解詳細的編譯指南。

### 快速編譯

```bash
# 安裝相依套件
make deps

# 編譯當前平台
make build

# 編譯所有平台
make build-all

# 執行測試
make test
```

## 📁 目錄結構

```
.
├── config.json          # 設定檔
├── data.db              # SQLite 資料庫 (自動產生)
├── views/               # HTML 模板
│   ├── index.html       # 請假申請頁面（立即提交 / 保存資料）
│   ├── schedule.html    # 排程管理頁面
│   └── result.html      # 結果頁面
├── config/              # 設定模組
├── controllers/         # 路由控制器
│   ├── form_controller.go
│   └── schedule_controller.go
└── models/              # 資料模型
    ├── leave_request.go
    ├── storage.go
    └── scheduler.go
```

## 📄 授權

MIT License

## ⚠️ 免責聲明

本工具僅供學習和個人使用。請確保您的使用方式符合相關服務條款和法律規定。
