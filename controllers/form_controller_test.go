package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"google-form-submitter/config"
	"google-form-submitter/models"
)

// 測試用配置
func setupTestConfig() *config.Config {
	return &config.Config{
		Port:    "8080",
		FormURL: "https://docs.google.com/forms/d/e/test/formResponse",
		EntryMap: map[string]string{
			"name":        "entry.123",
			"employee_id": "entry.456",
			"start_date":  "entry.789",
			"end_date":    "entry.012",
			"leave_type":  "entry.345",
			"password":    "entry.678",
		},
		DBPath: "",
	}
}

// 設定測試環境
func setupTestRouter(t *testing.T) (*gin.Engine, *FormController, *models.Storage, func()) {
	gin.SetMode(gin.TestMode)

	// 建立臨時資料庫
	tmpFile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("無法建立臨時資料庫: %v", err)
	}
	tmpFile.Close()

	storage, err := models.NewStorage(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("無法初始化 Storage: %v", err)
	}

	cfg := setupTestConfig()
	controller := NewFormController(cfg, storage)

	router := gin.New()
	router.LoadHTMLGlob("../views/*.html")

	// 註冊路由
	router.GET("/", controller.ShowForm)
	router.POST("/submit", controller.SubmitForm)
	router.POST("/api/submit", controller.SubmitAPI)
	router.GET("/api/saved", controller.ListSavedForms)
	router.POST("/api/saved", controller.SaveForm)
	router.GET("/api/saved/:id", controller.GetSavedForm)
	router.DELETE("/api/saved/:id", controller.DeleteSavedForm)

	cleanup := func() {
		storage.Close()
		os.Remove(tmpFile.Name())
	}

	return router, controller, storage, cleanup
}

// ============================================
// 12.1 測試網頁介面流程
// ============================================

// TestShowForm 測試表單顯示
// Requirements: 3.1
func TestShowForm(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("預期狀態碼 %d，實際 %d", http.StatusOK, w.Code)
	}

	// 檢查回應包含表單元素
	body := w.Body.String()
	expectedElements := []string{
		"name",
		"employee_id",
		"start_date",
		"end_date",
		"leave_type",
		"password",
		"submit",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(body, elem) {
			t.Errorf("回應應包含 %s 元素", elem)
		}
	}
}

// TestSubmitFormValidation 測試表單提交驗證
// Requirements: 3.2, 4.4
func TestSubmitFormValidation(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// 測試空表單提交
	form := url.Values{}
	req, _ := http.NewRequest("POST", "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("空表單應回傳 400，實際 %d", w.Code)
	}
}

// TestSubmitFormWithValidData 測試有效資料的表單提交
// Requirements: 3.2, 4.4
func TestSubmitFormWithValidData(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	form := url.Values{
		"name":        {"測試員工"},
		"employee_id": {"A12345"},
		"start_date":  {"2026-02-01"},
		"end_date":    {"2026-02-03"},
		"leave_type":  {"近假"},
		"password":    {"testpass"},
	}

	req, _ := http.NewRequest("POST", "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 由於實際 Google Form 可能無法連線，接受 200 或 502
	if w.Code != http.StatusOK && w.Code != http.StatusBadGateway {
		t.Errorf("預期狀態碼 200 或 502，實際 %d", w.Code)
	}
}

// ============================================
// 12.2 測試 API 端點
// ============================================

// TestSubmitAPI 測試 POST /api/submit
// Requirements: 5.1, 5.2, 5.3
func TestSubmitAPI(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// 測試有效 JSON 請求
	reqBody := map[string]string{
		"name":        "測試員工",
		"employee_id": "A12345",
		"start_date":  "2026-02-01",
		"end_date":    "2026-02-03",
		"leave_type":  "近假",
		"password":    "testpass",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 驗證回應為 JSON 格式
	var result models.SubmitResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("回應應為有效 JSON: %v", err)
	}

	// 驗證回應包含必要欄位
	if result.Message == "" {
		t.Error("回應應包含 message 欄位")
	}
}

// TestSubmitAPIValidation 測試 API 驗證
// Requirements: 5.2, 5.3
func TestSubmitAPIValidation(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// 測試缺少必填欄位
	reqBody := map[string]string{
		"name": "測試員工",
		// 缺少其他必填欄位
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("缺少必填欄位應回傳 400，實際 %d", w.Code)
	}

	var result models.SubmitResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("回應應為有效 JSON: %v", err)
	}

	if result.Success {
		t.Error("驗證失敗時 success 應為 false")
	}
}

// TestSubmitAPIInvalidJSON 測試無效 JSON
// Requirements: 5.2, 5.3
func TestSubmitAPIInvalidJSON(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req, _ := http.NewRequest("POST", "/api/submit", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("無效 JSON 應回傳 400，實際 %d", w.Code)
	}
}

// TestSaveFormAPI 測試 POST /api/saved
// Requirements: 8.2
func TestSaveFormAPI(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	reqBody := map[string]string{
		"label":       "測試資料",
		"name":        "測試員工",
		"employee_id": "A12345",
		"start_date":  "2026-02-01",
		"end_date":    "2026-02-03",
		"leave_type":  "近假",
		"password":    "testpass",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/saved", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("儲存應回傳 200，實際 %d", w.Code)
	}

	var result SaveFormResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("回應應為有效 JSON: %v", err)
	}

	if !result.Success {
		t.Error("儲存成功時 success 應為 true")
	}

	if result.ID <= 0 {
		t.Error("儲存成功應回傳有效 ID")
	}
}

// TestListSavedFormsAPI 測試 GET /api/saved
// Requirements: 8.3
func TestListSavedFormsAPI(t *testing.T) {
	router, _, storage, cleanup := setupTestRouter(t)
	defer cleanup()

	// 先儲存一筆資料
	savedForm := &models.SavedForm{
		Label:      "測試資料",
		Name:       "測試員工",
		EmployeeID: "A12345",
		StartDate:  "2026-02-01",
		EndDate:    "2026-02-03",
		LeaveType:  "近假",
		Password:   "testpass",
	}
	storage.Save(savedForm)

	req, _ := http.NewRequest("GET", "/api/saved", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("列表應回傳 200，實際 %d", w.Code)
	}

	var result ListSavedFormsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("回應應為有效 JSON: %v", err)
	}

	if !result.Success {
		t.Error("列表成功時 success 應為 true")
	}

	if len(result.Data) == 0 {
		t.Error("應有至少一筆資料")
	}
}

// TestGetSavedFormAPI 測試 GET /api/saved/:id
// Requirements: 8.3
func TestGetSavedFormAPI(t *testing.T) {
	router, _, storage, cleanup := setupTestRouter(t)
	defer cleanup()

	// 先儲存一筆資料
	savedForm := &models.SavedForm{
		Label:      "測試資料",
		Name:       "測試員工",
		EmployeeID: "A12345",
		StartDate:  "2026-02-01",
		EndDate:    "2026-02-03",
		LeaveType:  "近假",
		Password:   "testpass",
	}
	id, _ := storage.Save(savedForm)

	req, _ := http.NewRequest("GET", "/api/saved/"+string(rune(id+'0')), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 使用正確的 ID 格式
	req2, _ := http.NewRequest("GET", "/api/saved/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("取得單筆應回傳 200，實際 %d", w2.Code)
	}

	var result GetSavedFormResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &result); err != nil {
		t.Errorf("回應應為有效 JSON: %v", err)
	}

	if !result.Success {
		t.Error("取得成功時 success 應為 true")
	}
}

// TestGetSavedFormNotFound 測試取得不存在的資料
// Requirements: 8.3
func TestGetSavedFormNotFound(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", "/api/saved/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("不存在的資料應回傳 404，實際 %d", w.Code)
	}
}

// TestDeleteSavedFormAPI 測試 DELETE /api/saved/:id
// Requirements: 8.4
func TestDeleteSavedFormAPI(t *testing.T) {
	router, _, storage, cleanup := setupTestRouter(t)
	defer cleanup()

	// 先儲存一筆資料
	savedForm := &models.SavedForm{
		Label:      "測試資料",
		Name:       "測試員工",
		EmployeeID: "A12345",
		StartDate:  "2026-02-01",
		EndDate:    "2026-02-03",
		LeaveType:  "近假",
		Password:   "testpass",
	}
	storage.Save(savedForm)

	req, _ := http.NewRequest("DELETE", "/api/saved/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("刪除應回傳 200，實際 %d", w.Code)
	}

	var result DeleteSavedFormResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("回應應為有效 JSON: %v", err)
	}

	if !result.Success {
		t.Error("刪除成功時 success 應為 true")
	}

	// 確認已刪除
	req2, _ := http.NewRequest("GET", "/api/saved/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("刪除後應回傳 404，實際 %d", w2.Code)
	}
}

// TestDeleteSavedFormNotFound 測試刪除不存在的資料
// Requirements: 8.4
func TestDeleteSavedFormNotFound(t *testing.T) {
	router, _, _, cleanup := setupTestRouter(t)
	defer cleanup()

	req, _ := http.NewRequest("DELETE", "/api/saved/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("刪除不存在的資料應回傳 404，實際 %d", w.Code)
	}
}
