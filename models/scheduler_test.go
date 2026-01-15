package models

import (
	"os"
	"testing"
	"time"
)

// ============================================
// 12.3 測試排程功能
// ============================================

// setupTestStorage 建立測試用 Storage
func setupTestStorage(t *testing.T) (*Storage, func()) {
	tmpFile, err := os.CreateTemp("", "test_scheduler_*.db")
	if err != nil {
		t.Fatalf("無法建立臨時資料庫: %v", err)
	}
	tmpFile.Close()

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("無法初始化 Storage: %v", err)
	}

	cleanup := func() {
		storage.Close()
		os.Remove(tmpFile.Name())
	}

	return storage, cleanup
}

// TestNewScheduler 測試排程器建立
// Requirements: 7.1
func TestNewScheduler(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	cfg := &ScheduleConfig{
		Enabled:        true,
		Date:           "2026-12-01",
		SavedFormID:    1,
		PrepareSeconds: 5,
		RetryCount:     3,
		RetryInterval:  100,
	}

	submitter := NewGoogleFormSubmitter(
		"https://docs.google.com/forms/d/e/test/formResponse",
		map[string]string{
			"name":        "entry.123",
			"employee_id": "entry.456",
			"start_date":  "entry.789",
			"end_date":    "entry.012",
			"leave_type":  "entry.345",
			"password":    "entry.678",
		},
	)

	scheduler := NewScheduler(cfg, submitter, storage)

	if scheduler == nil {
		t.Error("NewScheduler 應回傳非 nil 的 Scheduler")
	}

	if scheduler.config != cfg {
		t.Error("Scheduler 應持有正確的配置")
	}

	if scheduler.submitter != submitter {
		t.Error("Scheduler 應持有正確的 submitter")
	}

	if scheduler.storage != storage {
		t.Error("Scheduler 應持有正確的 storage")
	}
}

// TestSchedulerStartWithDisabled 測試停用排程時的啟動
// Requirements: 7.7
func TestSchedulerStartWithDisabled(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	cfg := &ScheduleConfig{
		Enabled: false,
	}

	submitter := NewGoogleFormSubmitter(
		"https://docs.google.com/forms/d/e/test/formResponse",
		map[string]string{},
	)

	scheduler := NewScheduler(cfg, submitter, storage)
	err := scheduler.Start()

	if err != nil {
		t.Errorf("停用排程時啟動不應回傳錯誤: %v", err)
	}

	if scheduler.IsRunning() {
		t.Error("停用排程時不應處於運行狀態")
	}
}

// TestSchedulerStartWithPastDate 測試過去日期的排程
// Requirements: 7.1
func TestSchedulerStartWithPastDate(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// 儲存測試資料
	savedForm := &SavedForm{
		Label:      "測試",
		Name:       "測試員工",
		EmployeeID: "A12345",
		StartDate:  "2026-02-01",
		EndDate:    "2026-02-03",
		LeaveType:  "近假",
		Password:   "testpass",
	}
	storage.Save(savedForm)

	cfg := &ScheduleConfig{
		Enabled:        true,
		Date:           "2020-01-01", // 過去的日期
		SavedFormID:    1,
		PrepareSeconds: 5,
		RetryCount:     3,
		RetryInterval:  100,
	}

	submitter := NewGoogleFormSubmitter(
		"https://docs.google.com/forms/d/e/test/formResponse",
		map[string]string{
			"name":        "entry.123",
			"employee_id": "entry.456",
			"start_date":  "entry.789",
			"end_date":    "entry.012",
			"leave_type":  "entry.345",
			"password":    "entry.678",
		},
	)

	scheduler := NewScheduler(cfg, submitter, storage)
	err := scheduler.Start()

	// 過去日期應該不會啟動排程，但不應回傳錯誤
	if err != nil {
		t.Errorf("過去日期不應回傳錯誤: %v", err)
	}
}

// TestSchedulerStartWithInvalidSavedFormID 測試無效的 SavedFormID
// Requirements: 7.1, 8.6
func TestSchedulerStartWithInvalidSavedFormID(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// 使用未來日期
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	cfg := &ScheduleConfig{
		Enabled:        true,
		Date:           futureDate,
		SavedFormID:    9999, // 不存在的 ID
		PrepareSeconds: 5,
		RetryCount:     3,
		RetryInterval:  100,
	}

	submitter := NewGoogleFormSubmitter(
		"https://docs.google.com/forms/d/e/test/formResponse",
		map[string]string{
			"name":        "entry.123",
			"employee_id": "entry.456",
			"start_date":  "entry.789",
			"end_date":    "entry.012",
			"leave_type":  "entry.345",
			"password":    "entry.678",
		},
	)

	scheduler := NewScheduler(cfg, submitter, storage)
	err := scheduler.Start()

	if err == nil {
		t.Error("無效的 SavedFormID 應回傳錯誤")
	}
}

// TestGetNextRunTime 測試取得下次執行時間
// Requirements: 7.8
func TestGetNextRunTime(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// 儲存測試資料
	savedForm := &SavedForm{
		Label:      "測試",
		Name:       "測試員工",
		EmployeeID: "A12345",
		StartDate:  "2026-02-01",
		EndDate:    "2026-02-03",
		LeaveType:  "近假",
		Password:   "testpass",
	}
	storage.Save(savedForm)

	// 使用未來日期
	futureDate := time.Now().AddDate(1, 0, 0)
	futureDateStr := futureDate.Format("2006-01-02")

	cfg := &ScheduleConfig{
		Enabled:        true,
		Date:           futureDateStr,
		SavedFormID:    1,
		PrepareSeconds: 5,
		RetryCount:     3,
		RetryInterval:  100,
	}

	submitter := NewGoogleFormSubmitter(
		"https://docs.google.com/forms/d/e/test/formResponse",
		map[string]string{
			"name":        "entry.123",
			"employee_id": "entry.456",
			"start_date":  "entry.789",
			"end_date":    "entry.012",
			"leave_type":  "entry.345",
			"password":    "entry.678",
		},
	)

	scheduler := NewScheduler(cfg, submitter, storage)
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("啟動排程器失敗: %v", err)
	}
	defer scheduler.Stop()

	nextRun := scheduler.GetNextRunTime()

	// 驗證下次執行時間
	expectedDate := futureDate.Format("2006-01-02")
	actualDate := nextRun.Format("2006-01-02")

	if actualDate != expectedDate {
		t.Errorf("下次執行日期應為 %s，實際 %s", expectedDate, actualDate)
	}

	// 驗證時間為 00:00:00
	if nextRun.Hour() != 0 || nextRun.Minute() != 0 || nextRun.Second() != 0 {
		t.Errorf("下次執行時間應為 00:00:00，實際 %s", nextRun.Format("15:04:05"))
	}
}

// TestSchedulerStop 測試停止排程器
// Requirements: 7.1
func TestSchedulerStop(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// 儲存測試資料
	savedForm := &SavedForm{
		Label:      "測試",
		Name:       "測試員工",
		EmployeeID: "A12345",
		StartDate:  "2026-02-01",
		EndDate:    "2026-02-03",
		LeaveType:  "近假",
		Password:   "testpass",
	}
	storage.Save(savedForm)

	// 使用未來日期
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	cfg := &ScheduleConfig{
		Enabled:        true,
		Date:           futureDate,
		SavedFormID:    1,
		PrepareSeconds: 5,
		RetryCount:     3,
		RetryInterval:  100,
	}

	submitter := NewGoogleFormSubmitter(
		"https://docs.google.com/forms/d/e/test/formResponse",
		map[string]string{
			"name":        "entry.123",
			"employee_id": "entry.456",
			"start_date":  "entry.789",
			"end_date":    "entry.012",
			"leave_type":  "entry.345",
			"password":    "entry.678",
		},
	)

	scheduler := NewScheduler(cfg, submitter, storage)
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("啟動排程器失敗: %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("啟動後應處於運行狀態")
	}

	scheduler.Stop()

	if scheduler.IsRunning() {
		t.Error("停止後不應處於運行狀態")
	}
}

// TestParseScheduleDate 測試日期解析
// Requirements: 7.3
func TestParseScheduleDate(t *testing.T) {
	tests := []struct {
		name      string
		dateStr   string
		wantErr   bool
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "有效日期",
			dateStr:   "2026-02-01",
			wantErr:   false,
			wantYear:  2026,
			wantMonth: time.February,
			wantDay:   1,
		},
		{
			name:      "有效日期 - 年底",
			dateStr:   "2026-12-31",
			wantErr:   false,
			wantYear:  2026,
			wantMonth: time.December,
			wantDay:   31,
		},
		{
			name:    "空字串",
			dateStr: "",
			wantErr: true,
		},
		{
			name:    "無效格式",
			dateStr: "01-02-2026",
			wantErr: true,
		},
		{
			name:    "無效格式 - 斜線",
			dateStr: "2026/02/01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseScheduleDate(tt.dateStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("預期錯誤但沒有回傳錯誤")
				}
				return
			}

			if err != nil {
				t.Errorf("不預期錯誤: %v", err)
				return
			}

			if result.Year() != tt.wantYear {
				t.Errorf("年份應為 %d，實際 %d", tt.wantYear, result.Year())
			}

			if result.Month() != tt.wantMonth {
				t.Errorf("月份應為 %v，實際 %v", tt.wantMonth, result.Month())
			}

			if result.Day() != tt.wantDay {
				t.Errorf("日期應為 %d，實際 %d", tt.wantDay, result.Day())
			}

			// 驗證時間為 00:00:00
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 {
				t.Errorf("時間應為 00:00:00，實際 %s", result.Format("15:04:05"))
			}
		})
	}
}

// TestIsRunning 測試運行狀態檢查
// Requirements: 7.1
func TestIsRunning(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	cfg := &ScheduleConfig{
		Enabled: false,
	}

	submitter := NewGoogleFormSubmitter(
		"https://docs.google.com/forms/d/e/test/formResponse",
		map[string]string{},
	)

	scheduler := NewScheduler(cfg, submitter, storage)

	// 初始狀態應為未運行
	if scheduler.IsRunning() {
		t.Error("初始狀態不應處於運行狀態")
	}
}
