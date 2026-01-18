package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// ScheduleConfig 排程配置
type ScheduleConfig struct {
	Enabled        bool   `json:"enabled"`
	Date           string `json:"date"`            // YYYY-MM-DD 格式
	SavedFormID    int64  `json:"saved_form_id"`   // 要提交的儲存資料 ID
	PrepareSeconds int    `json:"prepare_seconds"` // 提前準備秒數，預設 5
	RetryCount     int    `json:"retry_count"`     // 失敗重試次數，預設 3
	RetryInterval  int    `json:"retry_interval"`  // 重試間隔毫秒，預設 100
}

// Config 應用程式配置
type Config struct {
	Port     string            `json:"port"`
	FormURL  string            `json:"form_url"`
	EntryMap map[string]string `json:"entry_map"`
	DBPath   string            `json:"db_path"`
	Schedule ScheduleConfig    `json:"schedule"`
}

// DefaultConfig 返回預設配置
func DefaultConfig() *Config {
	return &Config{
		Port:    "8080",
		FormURL: "",
		EntryMap: map[string]string{
			"name":        "",
			"employee_id": "",
			"start_date":  "",
			"end_date":    "",
			"leave_type":  "",
			"password":    "",
		},
		DBPath: "data.db",
		Schedule: ScheduleConfig{
			Enabled:        false,
			Date:           "",
			SavedFormID:    0,
			PrepareSeconds: 5,
			RetryCount:     3,
			RetryInterval:  100,
		},
	}
}

// Load 從 config.json 載入配置，支援環境變數覆蓋
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// 嘗試從配置檔載入
	if configPath == "" {
		configPath = "config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("讀取配置檔失敗: %w", err)
		}
		// 配置檔不存在，使用預設值並顯示警告
		fmt.Println("警告: 配置檔不存在，使用預設值")
	} else {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析配置檔失敗: %w", err)
		}
	}

	// 環境變數覆蓋
	applyEnvOverrides(cfg)

	// 驗證必要配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyEnvOverrides 套用環境變數覆蓋
func applyEnvOverrides(cfg *Config) {
	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}

	if formURL := os.Getenv("FORM_URL"); formURL != "" {
		cfg.FormURL = formURL
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	if scheduleEnabled := os.Getenv("SCHEDULE_ENABLED"); scheduleEnabled != "" {
		cfg.Schedule.Enabled = scheduleEnabled == "true" || scheduleEnabled == "1"
	}

	if scheduleDate := os.Getenv("SCHEDULE_DATE"); scheduleDate != "" {
		cfg.Schedule.Date = scheduleDate
	}

	if savedFormID := os.Getenv("SCHEDULE_SAVED_FORM_ID"); savedFormID != "" {
		if id, err := strconv.ParseInt(savedFormID, 10, 64); err == nil {
			cfg.Schedule.SavedFormID = id
		}
	}

	if prepareSeconds := os.Getenv("SCHEDULE_PREPARE_SECONDS"); prepareSeconds != "" {
		if seconds, err := strconv.Atoi(prepareSeconds); err == nil {
			cfg.Schedule.PrepareSeconds = seconds
		}
	}
}

// Validate 驗證配置是否有效
func (c *Config) Validate() error {
	if c.FormURL == "" {
		return fmt.Errorf("配置錯誤: form_url 未設定")
	}

	if len(c.EntryMap) == 0 {
		return fmt.Errorf("配置錯誤: entry_map 為空")
	}

	// 檢查必要的 entry 欄位
	requiredEntries := []string{"name", "employee_id", "start_date", "end_date", "leave_type", "password"}
	for _, entry := range requiredEntries {
		if c.EntryMap[entry] == "" {
			return fmt.Errorf("配置錯誤: entry_map 缺少 %s 欄位", entry)
		}
	}

	return nil
}
