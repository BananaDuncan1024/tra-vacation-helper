package models

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GoogleFormSubmitter Google Form 提交器
type GoogleFormSubmitter struct {
	FormURL    string
	EntryMap   map[string]string // 欄位名稱 → entry ID 對應
	HTTPClient *http.Client
}

// NewGoogleFormSubmitter 建立新的 GoogleFormSubmitter
func NewGoogleFormSubmitter(formURL string, entryMap map[string]string) *GoogleFormSubmitter {
	return &GoogleFormSubmitter{
		FormURL:  formURL,
		EntryMap: entryMap,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BuildFormData 建構 form-urlencoded 資料（公開供測試與排程使用）
func (s *GoogleFormSubmitter) BuildFormData(req *LeaveRequest) url.Values {
	data := url.Values{}

	if entryID, ok := s.EntryMap["name"]; ok && entryID != "" {
		data.Set(entryID, req.Name)
	}
	if entryID, ok := s.EntryMap["employee_id"]; ok && entryID != "" {
		data.Set(entryID, req.EmployeeID)
	}
	if entryID, ok := s.EntryMap["start_date"]; ok && entryID != "" {
		data.Set(entryID, req.StartDate)
	}
	if entryID, ok := s.EntryMap["end_date"]; ok && entryID != "" {
		data.Set(entryID, req.EndDate)
	}
	if entryID, ok := s.EntryMap["leave_type"]; ok && entryID != "" {
		data.Set(entryID, req.LeaveType)
	}
	if entryID, ok := s.EntryMap["password"]; ok && entryID != "" {
		data.Set(entryID, req.Password)
	}

	return data
}

// Submit 提交資料到 Google Form
func (s *GoogleFormSubmitter) Submit(req *LeaveRequest) (*SubmitResult, error) {
	// 驗證請求
	if err := Validate(req); err != nil {
		return &SubmitResult{
			Success: false,
			Message: fmt.Sprintf("驗證失敗：%s", err.Error()),
		}, nil
	}

	// 建構表單資料
	formData := s.BuildFormData(req)

	// 建立 HTTP 請求
	httpReq, err := http.NewRequest(
		"POST",
		s.FormURL,
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("建立請求失敗: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 發送請求
	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return &SubmitResult{
			Success: false,
			Message: "無法連線到 Google Form",
		}, nil
	}
	defer resp.Body.Close()

	// 讀取回應內容（用於除錯）
	_, _ = io.ReadAll(resp.Body)

	// Google Form 提交成功通常回傳 200
	if resp.StatusCode == http.StatusOK {
		return &SubmitResult{
			Success: true,
			Message: "表單提交成功",
		}, nil
	}

	return &SubmitResult{
		Success: false,
		Message: "Google Form 提交失敗",
	}, nil
}
