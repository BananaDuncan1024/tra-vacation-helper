package models

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// ScheduleConfig 排程配置（從 config 包複製以避免循環依賴）
type ScheduleConfig struct {
	Enabled        bool   `json:"enabled"`
	Date           string `json:"date"`            // YYYY-MM-DD 格式
	SavedFormID    int64  `json:"saved_form_id"`   // 要提交的儲存資料 ID
	PrepareSeconds int    `json:"prepare_seconds"` // 提前準備秒數，預設 5
	RetryCount     int    `json:"retry_count"`     // 失敗重試次數，預設 3
	RetryInterval  int    `json:"retry_interval"`  // 重試間隔毫秒，預設 100
}

// preparedRequest 預先準備的 HTTP 請求
type preparedRequest struct {
	formData   url.Values
	httpClient *http.Client
	targetURL  string
	request    *http.Request
}

// Scheduler 定時排程器
type Scheduler struct {
	config     *ScheduleConfig
	submitter  *GoogleFormSubmitter
	storage    *Storage
	cron       *cron.Cron
	logger     *log.Logger
	stopChan   chan struct{}
	mu         sync.Mutex
	running    bool
	targetTime time.Time
}

// NewScheduler 建立排程器
func NewScheduler(cfg *ScheduleConfig, submitter *GoogleFormSubmitter, storage *Storage) *Scheduler {
	return &Scheduler{
		config:    cfg,
		submitter: submitter,
		storage:   storage,
		logger:    log.New(os.Stdout, "[Scheduler] ", log.LstdFlags|log.Lmicroseconds),
		stopChan:  make(chan struct{}),
	}
}

// Start 啟動排程器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("排程器已在運行中")
	}

	if !s.config.Enabled {
		s.logger.Println("排程功能未啟用")
		return nil
	}

	// 解析排程日期
	targetTime, err := ParseScheduleDate(s.config.Date)
	if err != nil {
		return fmt.Errorf("排程日期格式錯誤: %w", err)
	}
	s.targetTime = targetTime

	// 檢查目標時間是否已過
	now := time.Now()
	if targetTime.Before(now) {
		s.logger.Printf("警告: 排程時間 %s 已過，排程將不會執行", targetTime.Format("2006-01-02 15:04:05"))
		return nil
	}

	// 驗證 SavedFormID 是否存在
	if s.config.SavedFormID <= 0 {
		return fmt.Errorf("排程配置錯誤: saved_form_id 未設定")
	}

	_, err = s.storage.GetByID(s.config.SavedFormID)
	if err != nil {
		return fmt.Errorf("排程配置錯誤: 找不到 ID 為 %d 的儲存資料", s.config.SavedFormID)
	}

	// 建立 cron 排程器
	s.cron = cron.New(cron.WithSeconds())

	// 計算準備時間（目標時間前 N 秒）
	prepareSeconds := s.config.PrepareSeconds
	if prepareSeconds <= 0 {
		prepareSeconds = 5
	}
	prepareTime := targetTime.Add(-time.Duration(prepareSeconds) * time.Second)

	// 如果準備時間已過但目標時間未過，直接進入準備狀態
	if prepareTime.Before(now) && targetTime.After(now) {
		s.logger.Println("準備時間已過，立即進入準備狀態")
		go s.executeWithPrecision()
	} else {
		// 設定在準備時間觸發
		cronSpec := fmt.Sprintf("%d %d %d %d %d *",
			prepareTime.Second(),
			prepareTime.Minute(),
			prepareTime.Hour(),
			prepareTime.Day(),
			int(prepareTime.Month()),
		)

		_, err = s.cron.AddFunc(cronSpec, func() {
			s.executeWithPrecision()
		})
		if err != nil {
			return fmt.Errorf("設定排程失敗: %w", err)
		}

		s.cron.Start()
	}

	s.running = true
	s.logger.Printf("排程器已啟動，目標時間: %s", targetTime.Format("2006-01-02 15:04:05.000"))

	return nil
}

// Stop 停止排程器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	// 發送停止信號
	close(s.stopChan)

	// 停止 cron
	if s.cron != nil {
		s.cron.Stop()
	}

	s.running = false
	s.logger.Println("排程器已停止")
}

// GetNextRunTime 取得下次執行時間
func (s *Scheduler) GetNextRunTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.targetTime
}

// calculateTargetTime 計算目標時間（排程日期的 00:00:00）
func (s *Scheduler) calculateTargetTime() time.Time {
	return s.targetTime
}

// prepareSubmission 準備提交（預先建立連線、構建資料）
func (s *Scheduler) prepareSubmission() (*preparedRequest, error) {
	// 從 Storage 讀取表單資料
	savedForm, err := s.storage.GetByID(s.config.SavedFormID)
	if err != nil {
		return nil, fmt.Errorf("讀取儲存資料失敗: %w", err)
	}

	// 轉換為 LeaveRequest
	req := savedForm.ToLeaveRequest()

	// 建構表單資料
	formData := s.submitter.BuildFormData(req)

	// 預先建立 HTTP 請求
	httpReq, err := http.NewRequest(
		"POST",
		s.submitter.FormURL,
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("建立 HTTP 請求失敗: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return &preparedRequest{
		formData:   formData,
		httpClient: s.submitter.HTTPClient,
		targetURL:  s.submitter.FormURL,
		request:    httpReq,
	}, nil
}

// executeWithPrecision 精確時間執行提交
func (s *Scheduler) executeWithPrecision() {
	s.logger.Println("進入準備狀態...")

	// 1. 計算目標時間
	targetTime := s.calculateTargetTime()

	// 2. 準備階段：預先建立連線、構建資料
	prepared, err := s.prepareSubmission()
	if err != nil {
		s.logger.Printf("準備失敗: %v", err)
		return
	}
	s.logger.Println("表單資料已準備完成")

	// 3. 計算等待時間
	waitDuration := time.Until(targetTime)
	if waitDuration < 0 {
		s.logger.Println("目標時間已過，立即執行")
		waitDuration = 0
	}

	s.logger.Printf("等待 %v 後執行提交...", waitDuration)

	// 4. 使用 time.NewTimer 精確等待到目標時間
	if waitDuration > 0 {
		timer := time.NewTimer(waitDuration)
		select {
		case <-timer.C:
			// 時間到，繼續執行
		case <-s.stopChan:
			timer.Stop()
			s.logger.Println("排程被取消")
			return
		}
	}

	// 5. 記錄實際執行時間
	actualTime := time.Now()
	s.logger.Printf("開始執行提交，實際時間: %s", actualTime.Format("2006-01-02 15:04:05.000"))

	// 6. 立即發送請求（帶重試）
	err = s.submitWithRetry(prepared)
	if err != nil {
		s.logger.Printf("提交失敗: %v", err)
	} else {
		s.logger.Printf("提交成功，耗時: %v", time.Since(actualTime))
	}
}

// submitWithRetry 帶重試的提交
func (s *Scheduler) submitWithRetry(prepared *preparedRequest) error {
	retryCount := s.config.RetryCount
	if retryCount <= 0 {
		retryCount = 3
	}

	retryInterval := s.config.RetryInterval
	if retryInterval <= 0 {
		retryInterval = 100
	}

	var lastErr error

	for i := 0; i < retryCount; i++ {
		if i > 0 {
			s.logger.Printf("第 %d 次重試...", i)
			time.Sleep(time.Duration(retryInterval) * time.Millisecond)

			// 重新建立請求（因為 Body 已被讀取）
			newReq, err := http.NewRequest(
				"POST",
				prepared.targetURL,
				strings.NewReader(prepared.formData.Encode()),
			)
			if err != nil {
				lastErr = fmt.Errorf("重建請求失敗: %w", err)
				continue
			}
			newReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			prepared.request = newReq
		}

		// 發送請求
		resp, err := prepared.httpClient.Do(prepared.request)
		if err != nil {
			lastErr = fmt.Errorf("發送請求失敗: %w", err)
			s.logger.Printf("請求失敗: %v", err)
			continue
		}
		resp.Body.Close()

		// 檢查回應狀態
		if resp.StatusCode == http.StatusOK {
			s.logger.Printf("Google Form 回應成功 (HTTP %d)", resp.StatusCode)
			return nil
		}

		lastErr = fmt.Errorf("Google Form 回應錯誤: HTTP %d", resp.StatusCode)
		s.logger.Printf("回應錯誤: HTTP %d", resp.StatusCode)
	}

	return fmt.Errorf("提交失敗，已重試 %d 次: %w", retryCount, lastErr)
}

// ParseScheduleDate 解析排程日期（YYYY-MM-DD 格式，時間固定為 00:00:00）
func ParseScheduleDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("日期字串為空")
	}

	// 解析 YYYY-MM-DD 格式
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("日期格式錯誤，請使用 YYYY-MM-DD 格式: %w", err)
	}

	return t, nil
}

// IsRunning 檢查排程器是否正在運行
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
