package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"google-form-submitter/models"
)

// ScheduleController 排程控制器
type ScheduleController struct {
	scheduler *models.Scheduler
	storage   *models.Storage
}

// NewScheduleController 建立新的 ScheduleController
func NewScheduleController(scheduler *models.Scheduler, storage *models.Storage) *ScheduleController {
	return &ScheduleController{
		scheduler: scheduler,
		storage:   storage,
	}
}

// ScheduleStatusResponse 排程狀態回應
type ScheduleStatusResponse struct {
	Success bool                   `json:"success"`
	Running bool                   `json:"running"`
	Config  *models.ScheduleConfig `json:"config,omitempty"`
	NextRun string                 `json:"next_run,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// CreateScheduleRequest 建立排程請求
type CreateScheduleRequest struct {
	Date           string `json:"date" binding:"required"`
	SavedFormID    int64  `json:"saved_form_id" binding:"required"`
	PrepareSeconds int    `json:"prepare_seconds"`
	RetryCount     int    `json:"retry_count"`
	RetryInterval  int    `json:"retry_interval"`
}

// ShowSchedule 顯示排程管理頁面
// GET /schedule
func (sc *ScheduleController) ShowSchedule(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "schedule.html", nil)
}

// GetScheduleStatus 取得排程狀態
// GET /api/schedule
func (sc *ScheduleController) GetScheduleStatus(ctx *gin.Context) {
	if sc.scheduler == nil {
		ctx.JSON(http.StatusOK, ScheduleStatusResponse{
			Success: true,
			Running: false,
			Message: "排程器未初始化",
		})
		return
	}

	running := sc.scheduler.IsRunning()
	resp := ScheduleStatusResponse{
		Success: true,
		Running: running,
		Config:  sc.scheduler.GetConfig(),
	}

	if running {
		nextRun := sc.scheduler.GetNextRunTime()
		resp.NextRun = nextRun.Format("2006-01-02 15:04:05")
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateSchedule 建立並啟動排程
// POST /api/schedule
func (sc *ScheduleController) CreateSchedule(ctx *gin.Context) {
	var req CreateScheduleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ScheduleStatusResponse{
			Success: false,
			Message: "請求格式錯誤或缺少必填欄位",
		})
		return
	}

	// 驗證 saved_form_id 存在
	_, err := sc.storage.GetByID(req.SavedFormID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ScheduleStatusResponse{
			Success: false,
			Message: "找不到指定的儲存資料",
		})
		return
	}

	// 設定預設值
	prepareSeconds := req.PrepareSeconds
	if prepareSeconds <= 0 {
		prepareSeconds = 5
	}
	retryCount := req.RetryCount
	if retryCount <= 0 {
		retryCount = 3
	}
	retryInterval := req.RetryInterval
	if retryInterval <= 0 {
		retryInterval = 100
	}

	// 建立排程配置
	cfg := &models.ScheduleConfig{
		Enabled:        true,
		Date:           req.Date,
		SavedFormID:    req.SavedFormID,
		PrepareSeconds: prepareSeconds,
		RetryCount:     retryCount,
		RetryInterval:  retryInterval,
	}

	// 確保 scheduler 已初始化
	if sc.scheduler == nil {
		ctx.JSON(http.StatusInternalServerError, ScheduleStatusResponse{
			Success: false,
			Message: "排程器未初始化",
		})
		return
	}

	// 啟動排程
	if err := sc.scheduler.StartWithConfig(cfg); err != nil {
		ctx.JSON(http.StatusBadRequest, ScheduleStatusResponse{
			Success: false,
			Message: "排程啟動失敗: " + err.Error(),
		})
		return
	}

	resp := ScheduleStatusResponse{
		Success: true,
		Running: true,
		Config:  cfg,
		Message: "排程已啟動",
	}

	nextRun := sc.scheduler.GetNextRunTime()
	resp.NextRun = nextRun.Format("2006-01-02 15:04:05")

	ctx.JSON(http.StatusOK, resp)
}

// StopSchedule 停止排程
// DELETE /api/schedule
func (sc *ScheduleController) StopSchedule(ctx *gin.Context) {
	if sc.scheduler == nil {
		ctx.JSON(http.StatusOK, ScheduleStatusResponse{
			Success: true,
			Running: false,
			Message: "排程器未初始化",
		})
		return
	}

	sc.scheduler.Stop()

	ctx.JSON(http.StatusOK, ScheduleStatusResponse{
		Success: true,
		Running: false,
		Message: "排程已停止",
	})
}
