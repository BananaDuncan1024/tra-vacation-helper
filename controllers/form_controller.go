package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"google-form-submitter/config"
	"google-form-submitter/models"
)

// FormController 表單控制器
type FormController struct {
	submitter *models.GoogleFormSubmitter
	storage   *models.Storage
	config    *config.Config
}

// NewFormController 建立新的 FormController
func NewFormController(cfg *config.Config, storage *models.Storage) *FormController {
	submitter := models.NewGoogleFormSubmitter(cfg.FormURL, cfg.EntryMap)
	return &FormController{
		submitter: submitter,
		storage:   storage,
		config:    cfg,
	}
}

// ShowForm 顯示表單頁面
// GET /
func (c *FormController) ShowForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", nil)
}

// SubmitForm 處理網頁表單提交
// POST /submit
func (c *FormController) SubmitForm(ctx *gin.Context) {
	var req models.LeaveRequest

	// 綁定表單資料
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.HTML(http.StatusBadRequest, "result.html", models.SubmitResult{
			Success: false,
			Message: "表單資料格式錯誤",
		})
		return
	}

	// 驗證表單資料
	if err := models.Validate(&req); err != nil {
		ctx.HTML(http.StatusBadRequest, "result.html", models.SubmitResult{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 提交到 Google Form
	result, err := c.submitter.Submit(&req)
	if err != nil {
		ctx.HTML(http.StatusBadGateway, "result.html", models.SubmitResult{
			Success: false,
			Message: "系統錯誤：" + err.Error(),
		})
		return
	}

	// 根據結果設定 HTTP 狀態碼
	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadGateway
	}

	ctx.HTML(statusCode, "result.html", result)
}

// SubmitAPI 處理 API JSON 提交
// POST /api/submit
func (c *FormController) SubmitAPI(ctx *gin.Context) {
	var req models.LeaveRequest

	// 綁定 JSON 資料
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.SubmitResult{
			Success: false,
			Message: "JSON 格式錯誤",
		})
		return
	}

	// 驗證表單資料
	if err := models.Validate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.SubmitResult{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 提交到 Google Form
	result, err := c.submitter.Submit(&req)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, models.SubmitResult{
			Success: false,
			Message: "系統錯誤：" + err.Error(),
		})
		return
	}

	// 根據結果設定 HTTP 狀態碼
	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusBadGateway
	}

	ctx.JSON(statusCode, result)
}

// SaveFormRequest 儲存表單請求結構
type SaveFormRequest struct {
	Label      string `json:"label" binding:"required"`
	Name       string `json:"name" binding:"required"`
	EmployeeID string `json:"employee_id" binding:"required"`
	StartDate  string `json:"start_date" binding:"required"`
	EndDate    string `json:"end_date" binding:"required"`
	LeaveType  string `json:"leave_type" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// SaveFormResponse 儲存表單回應結構
type SaveFormResponse struct {
	Success bool   `json:"success"`
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message"`
}

// ListSavedFormsResponse 列出已儲存表單回應結構
type ListSavedFormsResponse struct {
	Success bool                `json:"success"`
	Data    []*models.SavedForm `json:"data"`
}

// GetSavedFormResponse 取得單筆表單回應結構
type GetSavedFormResponse struct {
	Success bool              `json:"success"`
	Data    *models.SavedForm `json:"data,omitempty"`
	Message string            `json:"message,omitempty"`
}

// DeleteSavedFormResponse 刪除表單回應結構
type DeleteSavedFormResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SaveForm 儲存表單資料到 SQLite
// POST /api/saved
func (c *FormController) SaveForm(ctx *gin.Context) {
	var req SaveFormRequest

	// 綁定 JSON 資料
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, SaveFormResponse{
			Success: false,
			Message: "JSON 格式錯誤或缺少必填欄位",
		})
		return
	}

	// 建立 SavedForm
	savedForm := &models.SavedForm{
		Label:      req.Label,
		Name:       req.Name,
		EmployeeID: req.EmployeeID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		LeaveType:  req.LeaveType,
		Password:   req.Password,
	}

	// 儲存到資料庫
	id, err := c.storage.Save(savedForm)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, SaveFormResponse{
			Success: false,
			Message: "資料儲存失敗",
		})
		return
	}

	ctx.JSON(http.StatusOK, SaveFormResponse{
		Success: true,
		ID:      id,
		Message: "資料儲存成功",
	})
}

// ListSavedForms 列出已儲存的表單
// GET /api/saved
func (c *FormController) ListSavedForms(ctx *gin.Context) {
	forms, err := c.storage.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ListSavedFormsResponse{
			Success: false,
			Data:    nil,
		})
		return
	}

	// 確保回傳空陣列而非 null
	if forms == nil {
		forms = []*models.SavedForm{}
	}

	ctx.JSON(http.StatusOK, ListSavedFormsResponse{
		Success: true,
		Data:    forms,
	})
}

// GetSavedForm 取得單筆儲存的表單
// GET /api/saved/:id
func (c *FormController) GetSavedForm(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, GetSavedFormResponse{
			Success: false,
			Message: "無效的 ID 格式",
		})
		return
	}

	form, err := c.storage.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, GetSavedFormResponse{
			Success: false,
			Message: "找不到指定的資料",
		})
		return
	}

	ctx.JSON(http.StatusOK, GetSavedFormResponse{
		Success: true,
		Data:    form,
	})
}

// DeleteSavedForm 刪除已儲存的表單
// DELETE /api/saved/:id
func (c *FormController) DeleteSavedForm(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, DeleteSavedFormResponse{
			Success: false,
			Message: "無效的 ID 格式",
		})
		return
	}

	err = c.storage.Delete(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, DeleteSavedFormResponse{
			Success: false,
			Message: "找不到指定的資料",
		})
		return
	}

	ctx.JSON(http.StatusOK, DeleteSavedFormResponse{
		Success: true,
		Message: "資料刪除成功",
	})
}
