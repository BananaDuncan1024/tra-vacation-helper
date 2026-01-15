package models

// LeaveRequest 請假申請資料結構
type LeaveRequest struct {
	Name       string `json:"name" form:"name" binding:"required"`
	EmployeeID string `json:"employee_id" form:"employee_id" binding:"required"`
	StartDate  string `json:"start_date" form:"start_date" binding:"required"`
	EndDate    string `json:"end_date" form:"end_date" binding:"required"`
	LeaveType  string `json:"leave_type" form:"leave_type" binding:"required"`
	Password   string `json:"password" form:"password" binding:"required"`
}

// SubmitResult 提交結果資料結構
type SubmitResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
