# Design Document: Google Form Submitter

## Overview

本系統是一個使用 Golang Gin 框架建構的本地端 HTTP Server，採用 MVC 架構設計。系統提供網頁介面與 API 端點，讓使用者可以輸入請假資料並自動提交到指定的 Google Form。系統支援定時排程功能，可在指定日期午夜精確提交表單，並使用 SQLite 暫存表單資料。

### 技術選型

- **語言**: Go 1.21+
- **Web 框架**: Gin
- **模板引擎**: Gin 內建 HTML 模板
- **HTTP Client**: Go 標準庫 net/http
- **配置管理**: 環境變數 + JSON 配置檔
- **排程管理**: robfig/cron v3
- **資料庫**: SQLite (mattn/go-sqlite3)
- **測試框架**: testing/quick, gopter

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client (Browser/API)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Gin HTTP Server                          │
│                     (localhost:8080)                         │
└─────────────────────────────────────────────────────────────┘
                              │
    ┌─────────────────────────┼─────────────────────────┐
    ▼                         ▼                         ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   Controller    │ │      View       │ │     Config      │
│                 │ │                 │ │                 │
│ - FormHandler   │ │ - index.html    │ │ - config.json   │
│ - APIHandler    │ │ - result.html   │ │ - .env          │
│ - StorageAPI    │ │ - saved.html    │ │                 │
└─────────────────┘ └─────────────────┘ └─────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                         Model                                │
│                                                              │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │   Validator     │    │      GoogleFormSubmitter        │ │
│  │                 │    │                                 │ │
│  │ - ValidateForm  │───▶│ - Submit(data) → POST request  │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
│                                                              │
│  ┌─────────────────┐    ┌─────────────────────────────────┐ │
│  │   Storage       │    │         Scheduler               │ │
│  │   (SQLite)      │    │                                 │ │
│  │ - Save/Load     │◀──▶│ - CronJob                       │ │
│  │ - List/Delete   │    │ - PrecisionTimer                │ │
│  └─────────────────┘    └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Google Form Server                        │
│         (docs.google.com/forms/.../formResponse)            │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Main Entry Point (`main.go`)

```go
package main

func main() {
    // 載入配置
    cfg, err := config.Load()
    
    // 初始化 SQLite Storage
    storage, err := models.NewStorage(cfg.DBPath)
    
    // 初始化 Submitter
    submitter := models.NewGoogleFormSubmitter(cfg.FormURL, cfg.EntryMap)
    
    // 初始化 Scheduler（如果啟用）
    if cfg.Schedule.Enabled {
        scheduler := models.NewScheduler(&cfg.Schedule, submitter, storage)
        scheduler.Start()
    }
    
    // 初始化 Gin router
    // 註冊路由
    // 啟動 Server
}
```

### 2. Controller Layer (`controllers/`)

#### FormController

```go
type FormController struct {
    submitter *models.GoogleFormSubmitter
    storage   *models.Storage
    config    *config.Config
}

// ShowForm 顯示表單頁面
func (c *FormController) ShowForm(ctx *gin.Context)

// SubmitForm 處理表單提交（網頁）
func (c *FormController) SubmitForm(ctx *gin.Context)

// SubmitAPI 處理 API 提交（JSON）
func (c *FormController) SubmitAPI(ctx *gin.Context)

// SaveForm 儲存表單資料到 SQLite
func (c *FormController) SaveForm(ctx *gin.Context)

// ListSavedForms 列出已儲存的表單
func (c *FormController) ListSavedForms(ctx *gin.Context)

// DeleteSavedForm 刪除已儲存的表單
func (c *FormController) DeleteSavedForm(ctx *gin.Context)

// GetSavedForm 取得單筆儲存的表單
func (c *FormController) GetSavedForm(ctx *gin.Context)
```

### 3. Model Layer (`models/`)

#### LeaveRequest 資料結構

```go
type LeaveRequest struct {
    Name        string `json:"name" form:"name" binding:"required"`
    EmployeeID  string `json:"employee_id" form:"employee_id" binding:"required"`
    StartDate   string `json:"start_date" form:"start_date" binding:"required"`
    EndDate     string `json:"end_date" form:"end_date" binding:"required"`
    LeaveType   string `json:"leave_type" form:"leave_type" binding:"required"`
    Password    string `json:"password" form:"password" binding:"required"`
}
```

#### Validator

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// Validate 驗證 LeaveRequest
func Validate(req *LeaveRequest) []ValidationError

// validateRequired 驗證必填欄位
func validateRequired(req *LeaveRequest) []ValidationError

// validateDateFormat 驗證日期格式 (YYYY-MM-DD)
func validateDateFormat(date string) bool

// validateDateRange 驗證日期範圍（終點不早於起點）
func validateDateRange(start, end string) bool

// validateLeaveType 驗證假別（近假/長假）
func validateLeaveType(leaveType string) bool
```

#### GoogleFormSubmitter

```go
type GoogleFormSubmitter struct {
    FormURL     string
    EntryMap    map[string]string // 欄位名稱 → entry ID 對應
    HTTPClient  *http.Client
}

// NewGoogleFormSubmitter 建立 Submitter
func NewGoogleFormSubmitter(formURL string, entryMap map[string]string) *GoogleFormSubmitter

// Submit 提交資料到 Google Form
func (s *GoogleFormSubmitter) Submit(req *LeaveRequest) (*SubmitResult, error)

// BuildFormData 建構 form-urlencoded 資料（公開供測試）
func (s *GoogleFormSubmitter) BuildFormData(req *LeaveRequest) url.Values
```

#### SubmitResult

```go
type SubmitResult struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}
```

#### Storage (`models/storage.go`)

```go
type SavedForm struct {
    ID         int64     `json:"id"`
    Label      string    `json:"label"`       // 識別標籤
    Name       string    `json:"name"`
    EmployeeID string    `json:"employee_id"`
    StartDate  string    `json:"start_date"`
    EndDate    string    `json:"end_date"`
    LeaveType  string    `json:"leave_type"`
    Password   string    `json:"password"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type Storage struct {
    db *sql.DB
}

// NewStorage 建立 Storage 實例並初始化資料庫
func NewStorage(dbPath string) (*Storage, error)

// initDB 初始化資料庫表格
func (s *Storage) initDB() error

// Save 儲存表單資料
func (s *Storage) Save(form *SavedForm) (int64, error)

// GetByID 根據 ID 取得表單資料
func (s *Storage) GetByID(id int64) (*SavedForm, error)

// List 列出所有儲存的表單資料
func (s *Storage) List() ([]*SavedForm, error)

// Delete 刪除指定 ID 的表單資料
func (s *Storage) Delete(id int64) error

// Update 更新表單資料
func (s *Storage) Update(form *SavedForm) error

// ToLeaveRequest 轉換為 LeaveRequest
func (sf *SavedForm) ToLeaveRequest() *LeaveRequest
```

#### Scheduler (`models/scheduler.go`)

```go
type Scheduler struct {
    config     *config.ScheduleConfig
    submitter  *GoogleFormSubmitter
    storage    *Storage
    cron       *cron.Cron
    logger     *log.Logger
    stopChan   chan struct{}
}

type preparedRequest struct {
    formData   url.Values
    httpClient *http.Client
    targetURL  string
    request    *http.Request
}

// NewScheduler 建立排程器
func NewScheduler(cfg *config.ScheduleConfig, submitter *GoogleFormSubmitter, storage *Storage) *Scheduler

// Start 啟動排程器
func (s *Scheduler) Start() error

// Stop 停止排程器
func (s *Scheduler) Stop()

// GetNextRunTime 取得下次執行時間
func (s *Scheduler) GetNextRunTime() time.Time

// executeWithPrecision 精確時間執行提交
func (s *Scheduler) executeWithPrecision()

// prepareSubmission 提前準備提交（建立連線、構建資料）
func (s *Scheduler) prepareSubmission() (*preparedRequest, error)

// submitWithRetry 帶重試的提交
func (s *Scheduler) submitWithRetry(prepared *preparedRequest) error
```

#### 精確計時機制

```go
// executeWithPrecision 實作高精度計時
func (s *Scheduler) executeWithPrecision() {
    // 1. 計算目標時間（排程日期 00:00:00）
    targetTime := s.calculateTargetTime()
    
    // 2. 提前 N 秒進入準備狀態
    prepareTime := targetTime.Add(-time.Duration(s.config.PrepareSeconds) * time.Second)
    
    // 等待到準備時間
    select {
    case <-time.After(time.Until(prepareTime)):
    case <-s.stopChan:
        return
    }
    
    // 3. 準備階段：預先建立連線、構建資料
    prepared, err := s.prepareSubmission()
    if err != nil {
        s.logger.Printf("準備失敗: %v", err)
        return
    }
    
    // 4. 使用 time.NewTimer 精確等待到目標時間
    timer := time.NewTimer(time.Until(targetTime))
    select {
    case <-timer.C:
    case <-s.stopChan:
        timer.Stop()
        return
    }
    
    // 5. 立即發送請求（帶重試）
    s.submitWithRetry(prepared)
}

// prepareSubmission 準備提交
func (s *Scheduler) prepareSubmission() (*preparedRequest, error) {
    // 從 Storage 讀取表單資料
    savedForm, err := s.storage.GetByID(s.config.SavedFormID)
    if err != nil {
        return nil, err
    }
    
    // 轉換為 LeaveRequest
    req := savedForm.ToLeaveRequest()
    
    // 建構表單資料
    formData := s.submitter.BuildFormData(req)
    
    // 預先建立 HTTP 請求
    httpReq, err := http.NewRequest("POST", s.submitter.FormURL, 
        strings.NewReader(formData.Encode()))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    return &preparedRequest{
        formData:   formData,
        httpClient: s.submitter.HTTPClient,
        targetURL:  s.submitter.FormURL,
        request:    httpReq,
    }, nil
}
```

### 4. View Layer (`views/`)

#### index.html
- 表單輸入介面
- 包含所有請假欄位
- 表單驗證（前端）
- 儲存選項（checkbox）

#### result.html
- 顯示提交結果
- 成功/失敗訊息

#### saved.html
- 顯示已儲存的表單列表
- 提供刪除、載入功能

### 5. Config (`config/`)

```go
type Config struct {
    Port        string            `json:"port"`
    FormURL     string            `json:"form_url"`
    EntryMap    map[string]string `json:"entry_map"`
    DBPath      string            `json:"db_path"`
    Schedule    ScheduleConfig    `json:"schedule"`
}

type ScheduleConfig struct {
    Enabled        bool   `json:"enabled"`
    Date           string `json:"date"`            // YYYY-MM-DD 格式
    SavedFormID    int64  `json:"saved_form_id"`   // 要提交的儲存資料 ID
    PrepareSeconds int    `json:"prepare_seconds"` // 提前準備秒數，預設 5
    RetryCount     int    `json:"retry_count"`     // 失敗重試次數，預設 3
    RetryInterval  int    `json:"retry_interval"`  // 重試間隔毫秒，預設 100
}

// Load 從檔案或環境變數載入配置
func Load() (*Config, error)

// Validate 驗證配置完整性
func (c *Config) Validate() error

// ParseScheduleDate 解析排程日期
func ParseScheduleDate(dateStr string) (time.Time, error)
```

## Data Models

### LeaveRequest

| 欄位 | 類型 | 說明 | 驗證規則 |
|------|------|------|----------|
| Name | string | 員工姓名 | 必填 |
| EmployeeID | string | 員工代號 | 必填 |
| StartDate | string | 請假起點日期 (YYYY-MM-DD) | 必填、日期格式 |
| EndDate | string | 請假終點日期 (YYYY-MM-DD) | 必填、日期格式、不早於起點 |
| LeaveType | string | 假別 (近假/長假) | 必填、限定值 |
| Password | string | 請假密碼 | 必填 |

### SavedForm

| 欄位 | 類型 | 說明 |
|------|------|------|
| ID | int64 | 自動遞增主鍵 |
| Label | string | 識別標籤 |
| Name | string | 員工姓名 |
| EmployeeID | string | 員工代號 |
| StartDate | string | 請假起點日期 |
| EndDate | string | 請假終點日期 |
| LeaveType | string | 假別 |
| Password | string | 請假密碼 |
| CreatedAt | time.Time | 建立時間 |
| UpdatedAt | time.Time | 更新時間 |

### SQLite Schema

```sql
CREATE TABLE IF NOT EXISTS saved_forms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT NOT NULL,
    name TEXT NOT NULL,
    employee_id TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    leave_type TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_saved_forms_label ON saved_forms(label);
```

### Config (config.json)

```json
{
  "port": "8080",
  "form_url": "https://docs.google.com/forms/d/e/1FAIpQLSeVTwZeZDFmp0auR8ZkZ5-TvqYquV8Xorqc3J9MeS_liNacrw/formResponse",
  "entry_map": {
    "name": "entry.XXXXXXX",
    "employee_id": "entry.XXXXXXX",
    "start_date": "entry.XXXXXXX",
    "end_date": "entry.XXXXXXX",
    "leave_type": "entry.XXXXXXX",
    "password": "entry.XXXXXXX"
  },
  "db_path": "./data/forms.db",
  "schedule": {
    "enabled": false,
    "date": "2026-02-01",
    "saved_form_id": 1,
    "prepare_seconds": 5,
    "retry_count": 3,
    "retry_interval": 100
  }
}
```

## API Endpoints

| Method | Path | 說明 |
|--------|------|------|
| GET | / | 顯示表單頁面 |
| POST | /submit | 網頁表單提交 |
| POST | /api/submit | API JSON 提交 |
| GET | /api/saved | 列出已儲存的表單 |
| POST | /api/saved | 儲存表單資料 |
| GET | /api/saved/:id | 取得單筆儲存的表單 |
| DELETE | /api/saved/:id | 刪除已儲存的表單 |

### POST /api/submit

**Request Body:**
```json
{
  "name": "王小明",
  "employee_id": "A12345",
  "start_date": "2026-02-01",
  "end_date": "2026-02-03",
  "leave_type": "近假",
  "password": "mypassword",
  "save": true,
  "label": "二月請假"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "表單提交成功"
}
```

**Response (Error):**
```json
{
  "success": false,
  "message": "驗證失敗：姓名為必填欄位"
}
```

### POST /api/saved

**Request Body:**
```json
{
  "label": "常用請假資料",
  "name": "王小明",
  "employee_id": "A12345",
  "start_date": "2026-02-01",
  "end_date": "2026-02-03",
  "leave_type": "近假",
  "password": "mypassword"
}
```

**Response:**
```json
{
  "success": true,
  "id": 1,
  "message": "資料儲存成功"
}
```

### GET /api/saved

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "label": "常用請假資料",
      "name": "王小明",
      "employee_id": "A12345",
      "start_date": "2026-02-01",
      "end_date": "2026-02-03",
      "leave_type": "近假",
      "created_at": "2026-01-15T10:00:00Z"
    }
  ]
}
```



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Required Field Validation

*For any* LeaveRequest with one or more empty required fields (Name, EmployeeID, StartDate, EndDate, LeaveType, Password), the Validator SHALL return a validation error and the submission SHALL NOT proceed.

**Validates: Requirements 4.1, 4.2**

### Property 2: Form Data Construction

*For any* valid LeaveRequest and any EntryMap configuration, the BuildFormData function SHALL produce url.Values containing exactly the entry IDs from EntryMap as keys, mapped to their corresponding field values from the request.

**Validates: Requirements 4.3**

### Property 3: API Response Format

*For any* API request to POST /api/submit (valid or invalid), the response SHALL be valid JSON containing "success" (boolean) and "message" (string) fields.

**Validates: Requirements 5.2, 5.3**

### Property 4: Date Format Parsing

*For any* string in YYYY-MM-DD format where YYYY is 1900-2100, MM is 01-12, and DD is 01-31, the ParseScheduleDate function SHALL successfully parse it into a valid time.Time.

**Validates: Requirements 7.3**

### Property 5: Storage Round Trip

*For any* valid SavedForm with non-empty Label, saving to Storage then retrieving by ID SHALL return a SavedForm with identical Label, Name, EmployeeID, StartDate, EndDate, LeaveType, and Password values.

**Validates: Requirements 8.5**

### Property 6: Prepared Request Construction

*For any* valid SavedForm stored in SQLite, the prepareSubmission function SHALL produce a preparedRequest containing form data that matches the SavedForm's field values.

**Validates: Requirements 7.10**

## Error Handling

### Validation Errors

| 錯誤類型 | 錯誤訊息 | HTTP Status |
|---------|---------|-------------|
| 必填欄位為空 | "{欄位名稱}為必填欄位" | 400 Bad Request |
| 日期格式錯誤 | "日期格式錯誤，請使用 YYYY-MM-DD" | 400 Bad Request |
| 結束日期早於開始日期 | "請假終點日期不可早於起點日期" | 400 Bad Request |
| 假別無效 | "假別必須為「近假」或「長假」" | 400 Bad Request |

### Network Errors

| 錯誤類型 | 錯誤訊息 | HTTP Status |
|---------|---------|-------------|
| Google Form 連線失敗 | "無法連線到 Google Form" | 502 Bad Gateway |
| Google Form 回應錯誤 | "Google Form 提交失敗" | 502 Bad Gateway |
| 請求超時 | "請求超時，請稍後再試" | 504 Gateway Timeout |

### Storage Errors

| 錯誤類型 | 錯誤訊息 | HTTP Status |
|---------|---------|-------------|
| 資料庫連線失敗 | "無法連線到資料庫" | 500 Internal Server Error |
| 資料不存在 | "找不到指定的資料" | 404 Not Found |
| 儲存失敗 | "資料儲存失敗" | 500 Internal Server Error |

### Scheduler Errors

| 錯誤類型 | 處理方式 |
|---------|---------|
| 排程日期格式錯誤 | 啟動失敗並顯示錯誤 |
| 指定的 SavedFormID 不存在 | 記錄錯誤，排程不執行 |
| 提交失敗 | 根據 RetryCount 重試，記錄日誌 |

### Configuration Errors

| 錯誤類型 | 處理方式 |
|---------|---------|
| 配置檔不存在 | 使用預設值並顯示警告 |
| Form URL 未設定 | 啟動失敗並顯示錯誤 |
| Entry Map 為空 | 啟動失敗並顯示錯誤 |
| DB Path 未設定 | 使用預設值 ./data/forms.db |

## Testing Strategy

### Unit Tests

- **Validator Tests**: 測試各種驗證情境（空值、格式錯誤、邊界值）
- **Config Tests**: 測試配置載入與預設值
- **FormData Builder Tests**: 測試 URL 編碼資料建構
- **Storage Tests**: 測試 CRUD 操作
- **Scheduler Tests**: 測試排程邏輯（使用 mock 時間）

### Property-Based Tests

使用 Go 的 `gopter` 套件進行屬性測試：

1. **Property 1 (Required Field Validation)**: 生成隨機 LeaveRequest，隨機移除必填欄位，驗證回傳錯誤
2. **Property 2 (Form Data Construction)**: 生成隨機有效 LeaveRequest 和 EntryMap，驗證 BuildFormData 輸出包含所有欄位
3. **Property 3 (API Response Format)**: 生成隨機 API 請求，驗證回應為有效 JSON 且包含必要欄位
4. **Property 4 (Date Format Parsing)**: 生成隨機有效日期字串，驗證解析成功
5. **Property 5 (Storage Round Trip)**: 生成隨機 SavedForm，儲存後讀取，驗證資料一致
6. **Property 6 (Prepared Request Construction)**: 生成隨機 SavedForm，驗證 prepareSubmission 輸出正確

### Integration Tests

- **HTTP Handler Tests**: 使用 `httptest` 測試完整的 HTTP 請求/回應流程
- **Storage Integration Tests**: 測試 SQLite 實際操作
- **Scheduler Integration Tests**: 測試排程器與 Storage、Submitter 的整合

### Test Configuration

- 每個屬性測試至少執行 100 次迭代
- 使用 `gopter` 套件進行屬性測試
- 測試標籤格式: `Feature: google-form-submitter, Property N: {property_text}`

## Project Structure

```
google-form-submitter/
├── main.go                 # 程式進入點
├── config/
│   └── config.go          # 配置管理
├── controllers/
│   └── form_controller.go # HTTP 處理器
├── models/
│   ├── leave_request.go   # 資料結構
│   ├── validator.go       # 驗證邏輯
│   ├── submitter.go       # Google Form 提交
│   ├── storage.go         # SQLite 儲存
│   └── scheduler.go       # 定時排程
├── views/
│   ├── index.html         # 表單頁面
│   ├── result.html        # 結果頁面
│   └── saved.html         # 已儲存表單頁面
├── data/
│   └── forms.db           # SQLite 資料庫（自動建立）
├── config.json            # 配置檔
├── go.mod                 # Go 模組定義
└── go.sum                 # 依賴鎖定
```
