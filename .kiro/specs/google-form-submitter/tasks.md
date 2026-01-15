# Implementation Plan: Google Form Submitter

## Overview

使用 Golang Gin 框架建構本地端 HTTP Server，採用 MVC 架構，提供網頁介面與 API 端點提交資料到 Google Form。支援定時排程功能與 SQLite 資料暫存。

## Tasks

- [x] 1. 專案初始化與基礎設定
  - 建立 Go 模組 (go mod init)
  - 安裝 Gin 框架依賴
  - 安裝 SQLite 驅動 (mattn/go-sqlite3)
  - 安裝 Cron 套件 (robfig/cron/v3)
  - 建立專案目錄結構 (config/, controllers/, models/, views/, data/)
  - _Requirements: 1.1, 2.1_

- [x] 2. 配置管理模組
      - [x] 2.1 實作 config/config.go
    - 定義 Config 結構體（含 ScheduleConfig）
    - 實作 Load() 函數從 config.json 載入配置
    - 實作 Validate() 驗證配置完整性
    - 實作 ParseScheduleDate() 解析排程日期
    - 支援環境變數覆蓋
    - _Requirements: 6.1, 6.2, 6.3, 7.3, 7.7, 8.8_
  - [x] 2.2 更新 config.json 範本
    - 包含 port、form_url、entry_map、db_path、schedule 設定
    - _Requirements: 6.1, 6.2, 7.3, 7.4, 7.7_

- [x] 3. Model 層 - 基礎資料結構與驗證
  - [x] 3.1 實作 models/leave_request.go
    - 定義 LeaveRequest 結構體
    - 定義 SubmitResult 結構體
    - _Requirements: 4.1_
  - [x] 3.2 實作 models/validator.go
    - 實作 Validate() 函數驗證必填欄位
    - 實作日期格式驗證 (YYYY-MM-DD)
    - 實作日期邏輯驗證（終點不早於起點）
    - 實作假別驗證（近假/長假）
    - _Requirements: 4.1, 4.2_
  - [ ]* 3.3 撰寫 Validator 屬性測試
    - **Property 1: Required Field Validation**
    - **Validates: Requirements 4.1, 4.2**

- [x] 4. Model 層 - Google Form 提交
  - [x] 4.1 實作 models/submitter.go
    - 定義 GoogleFormSubmitter 結構體
    - 實作 NewGoogleFormSubmitter() 建構函數
    - 實作 BuildFormData() 建構 form-urlencoded 資料
    - 實作 Submit() 發送 POST 請求到 Google Form
    - _Requirements: 4.3, 4.4, 4.5_
  - [ ]* 4.2 撰寫 Submitter 屬性測試
    - **Property 2: Form Data Construction**
    - **Validates: Requirements 4.3**

- [x] 5. Model 層 - SQLite 儲存
  - [x] 5.1 實作 models/storage.go
    - 定義 SavedForm 結構體
    - 定義 Storage 結構體
    - 實作 NewStorage() 建立連線並初始化資料庫
    - 實作 initDB() 建立資料表
    - 實作 Save() 儲存表單資料
    - 實作 GetByID() 根據 ID 取得資料
    - 實作 List() 列出所有資料
    - 實作 Delete() 刪除資料
    - 實作 ToLeaveRequest() 轉換函數
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.7_
  - [ ]* 5.2 撰寫 Storage 屬性測試
    - **Property 5: Storage Round Trip**
    - **Validates: Requirements 8.5**

- [x] 6. Model 層 - 定時排程
  - [x] 6.1 實作 models/scheduler.go
    - 定義 Scheduler 結構體
    - 定義 preparedRequest 結構體
    - 實作 NewScheduler() 建構函數
    - 實作 Start() 啟動排程器
    - 實作 Stop() 停止排程器
    - 實作 GetNextRunTime() 取得下次執行時間
    - 實作 calculateTargetTime() 計算目標時間
    - 實作 prepareSubmission() 準備提交
    - 實作 executeWithPrecision() 精確時間執行
    - 實作 submitWithRetry() 帶重試的提交
    - _Requirements: 7.1, 7.2, 7.5, 7.6, 7.8, 7.9, 7.10, 7.11, 8.6_
  - [ ]* 6.2 撰寫 Scheduler 屬性測試
    - **Property 4: Date Format Parsing**
    - **Property 6: Prepared Request Construction**
    - **Validates: Requirements 7.3, 7.10**

- [x] 7. Checkpoint - 確認 Model 層完成
  - 確保所有 Model 層程式碼編譯通過
  - 確保資料庫可正常初始化
  - 如有問題請詢問使用者

- [x] 8. View 層實作
  - [x] 8.1 更新 views/index.html
    - 表單輸入介面（姓名、員工代號、日期、假別、密碼）
    - 新增儲存選項（checkbox + label 輸入）
    - 前端基本驗證
    - _Requirements: 3.1, 3.2, 8.2_
  - [x] 8.2 更新 views/result.html
    - 顯示提交結果（成功/失敗訊息）
    - _Requirements: 4.4, 4.5_
  - [ ] 8.3 建立 views/saved.html
    - 顯示已儲存的表單列表
    - 提供刪除、載入功能
    - _Requirements: 8.3, 8.4, 8.5_

- [x] 9. Controller 層實作
  - [x] 9.1 更新 controllers/form_controller.go - 基礎功能
    - 實作 ShowForm() 顯示表單頁面
    - 實作 SubmitForm() 處理網頁表單提交
    - 實作 SubmitAPI() 處理 API JSON 提交
    - _Requirements: 3.1, 4.2, 4.4, 4.5, 5.1, 5.2, 5.3_
  - [x] 9.2 更新 controllers/form_controller.go - 儲存功能
    - 實作 SaveForm() 儲存表單資料
    - 實作 ListSavedForms() 列出已儲存表單
    - 實作 GetSavedForm() 取得單筆表單
    - 實作 DeleteSavedForm() 刪除表單
    - _Requirements: 8.2, 8.3, 8.4, 8.5_
  - [ ]* 9.3 撰寫 Controller 屬性測試
    - **Property 3: API Response Format**
    - **Validates: Requirements 5.2, 5.3**

- [x] 10. 主程式整合
  - [x] 10.1 更新 main.go
    - 載入配置
    - 初始化 SQLite Storage
    - 初始化 GoogleFormSubmitter
    - 初始化 Scheduler（如果啟用）
    - 初始化 Gin router
    - 註冊所有路由
    - 顯示啟動訊息與排程資訊
    - 啟動 Server
    - _Requirements: 1.1, 1.2, 1.3, 5.1, 7.1, 7.8, 8.7_

- [x] 11. Checkpoint - 完整功能測試
  - 確保所有程式碼編譯通過
  - 確保 Server 可正常啟動
  - 確保 SQLite 資料庫自動建立
  - 如有問題請詢問使用者

- [x] 12. 整合測試
  - [x] 12.1 測試網頁介面流程
    - 測試表單顯示
    - 測試表單提交
    - 測試儲存功能
    - _Requirements: 3.1, 3.2, 4.4, 8.2_
  - [x] 12.2 測試 API 端點
    - 測試 POST /api/submit
    - 測試 GET/POST/DELETE /api/saved
    - _Requirements: 5.1, 5.2, 5.3, 8.3, 8.4_
  - [x] 12.3 測試排程功能
    - 測試排程器啟動
    - 測試 GetNextRunTime()
    - _Requirements: 7.1, 7.8_

- [-] 13. Final Checkpoint - 確保所有功能正常
  - 確保所有測試通過
  - 確保排程功能可正常運作
  - 如有問題請詢問使用者

## Notes

- 標記 `*` 的任務為選擇性任務，可跳過以加速 MVP 開發
- 每個任務都有對應的需求追溯
- Checkpoint 用於確保階段性驗證
- 屬性測試驗證通用正確性屬性
- 單元測試驗證特定範例與邊界情況
- 已完成的任務（標記 [x]）為先前實作的基礎功能
- 新增任務主要涵蓋 Requirement 7（排程）和 Requirement 8（SQLite 儲存）
