package models

import (
	"errors"
	"time"
)

// 允許的假別類型
var allowedLeaveTypes = map[string]bool{
	"近假": true,
	"長假": true,
}

// ValidationError 驗證錯誤
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Validate 驗證 LeaveRequest 的所有欄位
func Validate(req *LeaveRequest) error {
	// 驗證必填欄位
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "姓名為必填欄位"}
	}
	if req.EmployeeID == "" {
		return &ValidationError{Field: "employee_id", Message: "員工代號為必填欄位"}
	}
	if req.StartDate == "" {
		return &ValidationError{Field: "start_date", Message: "請假起點日期為必填欄位"}
	}
	if req.EndDate == "" {
		return &ValidationError{Field: "end_date", Message: "請假終點日期為必填欄位"}
	}
	if req.LeaveType == "" {
		return &ValidationError{Field: "leave_type", Message: "假別為必填欄位"}
	}
	if req.Password == "" {
		return &ValidationError{Field: "password", Message: "請假密碼為必填欄位"}
	}

	// 驗證日期格式
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return &ValidationError{Field: "start_date", Message: "日期格式錯誤，請使用 YYYY-MM-DD"}
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return &ValidationError{Field: "end_date", Message: "日期格式錯誤，請使用 YYYY-MM-DD"}
	}

	// 驗證日期邏輯（終點不早於起點）
	if endDate.Before(startDate) {
		return &ValidationError{Field: "end_date", Message: "請假終點日期不可早於起點日期"}
	}

	// 驗證假別
	if !allowedLeaveTypes[req.LeaveType] {
		return &ValidationError{Field: "leave_type", Message: "假別必須為「近假」或「長假」"}
	}

	return nil
}

// parseDate 解析日期字串 (YYYY-MM-DD 格式)
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, errors.New("日期不可為空")
	}
	return time.Parse("2006-01-02", dateStr)
}
