# Requirements Document

## Introduction

本功能是一個使用 Golang Gin 框架建構的本地端 HTTP Server，採用 MVC 架構設計，提供使用者透過 API 或網頁介面輸入參數，並將資料直接提交到 Google Form 表單（線上請假登記系統）。

## Target Google Form

- **Form URL**: `https://docs.google.com/forms/d/e/1FAIpQLSeVTwZeZDFmp0auR8ZkZ5-TvqYquV8Xorqc3J9MeS_liNacrw/viewform`
- **Submit URL**: `https://docs.google.com/forms/d/e/1FAIpQLSeVTwZeZDFmp0auR8ZkZ5-TvqYquV8Xorqc3J9MeS_liNacrw/formResponse`

### 表單欄位（需從表單 HTML 取得實際 entry ID）

| 欄位名稱 | 說明 | Entry ID |
|---------|------|----------|
| 姓名 | 員工姓名 | entry.XXXXXXX |
| 員工代號 | 員工識別碼 | entry.XXXXXXX |
| 請假起點日期 | 休假開始日期 | entry.XXXXXXX |
| 請假終點日期 | 休假結束日期 | entry.XXXXXXX |
| 假別 | 近假/長假 | entry.XXXXXXX |
| 請假密碼 | 驗證密碼 | entry.XXXXXXX |

> **注意**: 實際的 entry ID 需要從表單的 HTML 原始碼中取得，或透過瀏覽器開發者工具查看。

## Glossary

- **Server**: 本地端運行的 Gin HTTP 伺服器
- **Controller**: 處理 HTTP 請求並協調 Model 與 View 的元件
- **Model**: 負責資料處理與 Google Form 提交邏輯的元件
- **View**: 負責呈現使用者介面的元件（HTML 模板）
- **Google_Form**: Google 提供的線上表單服務
- **Form_Entry**: Google Form 中的欄位識別碼（entry.XXXXXXX 格式）
- **Scheduler**: 負責定時排程任務的元件
- **SQLite**: 輕量級本地資料庫，用於暫存表單資料
- **Saved_Form**: 儲存在 SQLite 中的表單資料記錄

## Requirements

### Requirement 1: Server 啟動與配置

**User Story:** As a 開發者, I want to 在本地端啟動一個 HTTP Server, so that 我可以透過瀏覽器或 API 存取服務。

#### Acceptance Criteria

1. WHEN Server 啟動時, THE Server SHALL 監聽指定的 port（預設 8080）
2. WHEN Server 成功啟動時, THE Server SHALL 在終端機顯示啟動訊息與存取網址
3. IF Server 啟動失敗（如 port 被佔用）, THEN THE Server SHALL 顯示錯誤訊息並終止程式

### Requirement 2: MVC 架構設計

**User Story:** As a 開發者, I want to 使用 MVC 架構組織程式碼, so that 程式碼易於維護與擴展。

#### Acceptance Criteria

1. THE Server SHALL 將程式碼分離為 Controller、Model、View 三個層級
2. THE Controller SHALL 負責處理 HTTP 路由與請求
3. THE Model SHALL 負責資料驗證與 Google Form 提交邏輯
4. THE View SHALL 負責呈現 HTML 表單介面

### Requirement 3: 表單輸入介面

**User Story:** As a 使用者, I want to 透過網頁介面輸入表單資料, so that 我可以方便地填寫要提交的內容。

#### Acceptance Criteria

1. WHEN 使用者訪問首頁時, THE View SHALL 顯示一個 HTML 表單介面
2. THE View SHALL 提供輸入欄位讓使用者填寫 Google Form 對應的資料
3. WHEN 使用者點擊提交按鈕時, THE View SHALL 將表單資料發送到 Server

### Requirement 4: Google Form 提交功能

**User Story:** As a 使用者, I want to 將輸入的資料提交到 Google Form, so that 資料可以被記錄到 Google 試算表。

#### Acceptance Criteria

1. WHEN 收到表單提交請求時, THE Model SHALL 驗證必填欄位是否已填寫
2. IF 驗證失敗, THEN THE Controller SHALL 回傳錯誤訊息給使用者
3. WHEN 驗證成功時, THE Model SHALL 將資料以 POST 請求發送到 Google Form 的提交網址
4. WHEN Google Form 回應成功時, THE Controller SHALL 回傳成功訊息給使用者
5. IF Google Form 回應失敗, THEN THE Controller SHALL 回傳錯誤訊息給使用者

### Requirement 5: API 端點支援

**User Story:** As a 開發者, I want to 透過 API 端點提交資料, so that 我可以整合到其他系統或自動化流程。

#### Acceptance Criteria

1. THE Server SHALL 提供 POST /api/submit 端點接收 JSON 格式的表單資料
2. WHEN 收到 API 請求時, THE Controller SHALL 解析 JSON 並執行提交流程
3. THE Controller SHALL 回傳 JSON 格式的回應（成功或失敗訊息）

### Requirement 6: 配置管理

**User Story:** As a 開發者, I want to 透過配置檔設定 Google Form 資訊, so that 我可以輕鬆切換不同的表單。

#### Acceptance Criteria

1. THE Server SHALL 支援透過環境變數或配置檔設定 Google Form URL
2. THE Server SHALL 支援配置 Form Entry ID 與欄位名稱的對應關係
3. IF 配置缺失, THEN THE Server SHALL 在啟動時顯示警告訊息

### Requirement 7: 定時排程提交（Cron Job）

**User Story:** As a 使用者, I want to 設定定時排程在指定日期的午夜 12 點自動提交表單, so that 我可以在搶票/預約開放的第一時間自動送出申請。

#### Acceptance Criteria

1. THE Server SHALL 支援 Cron Job 排程功能，在 Server 運行期間持續監控排程
2. WHEN 到達設定的排程日期午夜 00:00:00 整點時, THE Scheduler SHALL 自動觸發 Google Form 提交
3. THE Server SHALL 支援透過配置檔設定排程日期（格式：YYYY-MM-DD），時間固定為 00:00:00
4. THE Server SHALL 支援透過配置檔預設要提交的表單資料
5. WHEN 排程觸發提交時, THE Server SHALL 記錄提交時間與結果到日誌
6. IF 排程提交失敗, THEN THE Server SHALL 記錄錯誤訊息並可選擇重試
7. THE Server SHALL 支援啟用或停用排程功能的配置選項
8. WHEN Server 啟動且排程功能啟用時, THE Server SHALL 顯示下次排程執行時間
9. THE Scheduler SHALL 在排程時間前數秒（預設 5 秒）進入準備狀態，預先建立 HTTP 連線
10. WHEN 進入準備狀態時, THE Scheduler SHALL 預先構建表單資料並準備好 HTTP 請求
11. THE Scheduler SHALL 使用高精度計時器，確保在 00:00:00.000 時精確發送請求

### Requirement 8: 表單資料暫存（SQLite）

**User Story:** As a 使用者, I want to 將輸入的表單資料暫存到本地資料庫, so that 我可以保存常用資料、避免重複輸入，並供排程任務使用。

#### Acceptance Criteria

1. THE Server SHALL 使用 SQLite 作為本地資料庫儲存表單資料
2. WHEN 使用者透過網頁或 API 提交表單時, THE Server SHALL 提供選項將資料儲存到 SQLite
3. THE Server SHALL 提供 API 端點查詢已儲存的表單資料列表
4. THE Server SHALL 提供 API 端點刪除已儲存的表單資料
5. THE Server SHALL 支援為每筆儲存的資料設定名稱標籤，方便識別
6. WHEN 排程功能啟用時, THE Scheduler SHALL 從 SQLite 讀取指定的表單資料進行提交
7. THE Server SHALL 在啟動時自動建立 SQLite 資料庫檔案（若不存在）
8. THE Server SHALL 支援透過配置檔設定 SQLite 資料庫檔案路徑
