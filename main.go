package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"google-form-submitter/config"
	"google-form-submitter/controllers"
	"google-form-submitter/models"
)

func main() {
	// 載入配置
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("載入配置失敗: %v", err)
		os.Exit(1)
	}

	// 初始化 SQLite Storage
	storage, err := models.NewStorage(cfg.DBPath)
	if err != nil {
		log.Fatalf("初始化資料庫失敗: %v", err)
		os.Exit(1)
	}
	defer storage.Close()

	// 初始化 GoogleFormSubmitter
	submitter := models.NewGoogleFormSubmitter(cfg.FormURL, cfg.EntryMap)

	// 初始化 Scheduler（如果啟用）
	var scheduler *models.Scheduler
	if cfg.Schedule.Enabled {
		scheduleConfig := &models.ScheduleConfig{
			Enabled:        cfg.Schedule.Enabled,
			Date:           cfg.Schedule.Date,
			SavedFormID:    cfg.Schedule.SavedFormID,
			PrepareSeconds: cfg.Schedule.PrepareSeconds,
			RetryCount:     cfg.Schedule.RetryCount,
			RetryInterval:  cfg.Schedule.RetryInterval,
		}
		scheduler = models.NewScheduler(scheduleConfig, submitter, storage)
		if err := scheduler.Start(); err != nil {
			log.Printf("警告: 排程器啟動失敗: %v", err)
		}
	}

	// 初始化 Gin router
	router := gin.Default()

	// 載入 HTML 模板
	router.LoadHTMLGlob("views/*.html")

	// 建立 Controller
	formController := controllers.NewFormController(cfg, storage)

	// 註冊路由
	// GET / - 顯示表單頁面
	router.GET("/", formController.ShowForm)

	// POST /submit - 網頁表單提交
	router.POST("/submit", formController.SubmitForm)

	// POST /api/submit - API JSON 提交
	router.POST("/api/submit", formController.SubmitAPI)

	// Storage API 路由
	// GET /api/saved - 列出已儲存的表單
	router.GET("/api/saved", formController.ListSavedForms)

	// POST /api/saved - 儲存表單資料
	router.POST("/api/saved", formController.SaveForm)

	// GET /api/saved/:id - 取得單筆儲存的表單
	router.GET("/api/saved/:id", formController.GetSavedForm)

	// DELETE /api/saved/:id - 刪除已儲存的表單
	router.DELETE("/api/saved/:id", formController.DeleteSavedForm)

	// 顯示啟動訊息
	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Println("========================================")
	fmt.Println("  Google Form Submitter Server")
	fmt.Println("========================================")
	fmt.Printf("存取網址: http://localhost%s\n", addr)
	fmt.Printf("資料庫路徑: %s\n", cfg.DBPath)

	// 顯示排程資訊
	if cfg.Schedule.Enabled && scheduler != nil && scheduler.IsRunning() {
		nextRun := scheduler.GetNextRunTime()
		fmt.Println("----------------------------------------")
		fmt.Println("排程功能: 已啟用")
		fmt.Printf("排程日期: %s\n", cfg.Schedule.Date)
		fmt.Printf("下次執行時間: %s\n", nextRun.Format("2006-01-02 15:04:05"))
		fmt.Printf("使用儲存資料 ID: %d\n", cfg.Schedule.SavedFormID)
		fmt.Printf("提前準備秒數: %d\n", cfg.Schedule.PrepareSeconds)
		fmt.Printf("失敗重試次數: %d\n", cfg.Schedule.RetryCount)
	} else if cfg.Schedule.Enabled {
		fmt.Println("----------------------------------------")
		fmt.Println("排程功能: 已設定但未啟動（可能時間已過或配置錯誤）")
	} else {
		fmt.Println("----------------------------------------")
		fmt.Println("排程功能: 未啟用")
	}
	fmt.Println("========================================")

	// 設定優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在背景啟動 Server
	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("Server 啟動失敗: %v", err)
		}
	}()

	fmt.Println("Server 已啟動，按 Ctrl+C 停止...")

	// 等待關閉信號
	<-quit
	fmt.Println("\n正在關閉 Server...")

	// 停止排程器
	if scheduler != nil {
		scheduler.Stop()
	}

	fmt.Println("Server 已關閉")
}
