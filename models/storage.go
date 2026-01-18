package models

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SavedForm 儲存在 SQLite 中的表單資料記錄
type SavedForm struct {
	ID         int64     `json:"id"`
	Label      string    `json:"label"` // 識別標籤
	Name       string    `json:"name"`
	EmployeeID string    `json:"employee_id"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	LeaveType  string    `json:"leave_type"`
	Password   string    `json:"password"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ToLeaveRequest 轉換為 LeaveRequest
func (sf *SavedForm) ToLeaveRequest() *LeaveRequest {
	return &LeaveRequest{
		Name:       sf.Name,
		EmployeeID: sf.EmployeeID,
		StartDate:  sf.StartDate,
		EndDate:    sf.EndDate,
		LeaveType:  sf.LeaveType,
		Password:   sf.Password,
	}
}

// Storage SQLite 儲存管理器
type Storage struct {
	db *sql.DB
}

// NewStorage 建立 Storage 實例並初始化資料庫
func NewStorage(dbPath string) (*Storage, error) {
	// 確保目錄存在
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("無法建立資料庫目錄: %w", err)
		}
	}

	// 開啟資料庫連線
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("無法連線到資料庫: %w", err)
	}

	// 測試連線
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("無法連線到資料庫: %w", err)
	}

	storage := &Storage{db: db}

	// 初始化資料表
	if err := storage.initDB(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// initDB 初始化資料庫表格
func (s *Storage) initDB() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS saved_forms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		label TEXT NOT NULL,
		name TEXT NOT NULL,
		employee_id TEXT NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		leave_type TEXT NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_saved_forms_label ON saved_forms(label);
	`

	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("初始化資料庫失敗: %w", err)
	}

	return nil
}

// Save 儲存表單資料
func (s *Storage) Save(form *SavedForm) (int64, error) {
	now := time.Now()

	result, err := s.db.Exec(`
		INSERT INTO saved_forms (label, name, employee_id, start_date, end_date, leave_type, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, form.Label, form.Name, form.EmployeeID, form.StartDate, form.EndDate, form.LeaveType, form.Password, now, now)

	if err != nil {
		return 0, fmt.Errorf("資料儲存失敗: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("取得 ID 失敗: %w", err)
	}

	return id, nil
}

// GetByID 根據 ID 取得表單資料
func (s *Storage) GetByID(id int64) (*SavedForm, error) {
	row := s.db.QueryRow(`
		SELECT id, label, name, employee_id, start_date, end_date, leave_type, password, created_at, updated_at
		FROM saved_forms
		WHERE id = ?
	`, id)

	form := &SavedForm{}
	err := row.Scan(
		&form.ID,
		&form.Label,
		&form.Name,
		&form.EmployeeID,
		&form.StartDate,
		&form.EndDate,
		&form.LeaveType,
		&form.Password,
		&form.CreatedAt,
		&form.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("找不到指定的資料")
	}
	if err != nil {
		return nil, fmt.Errorf("查詢資料失敗: %w", err)
	}

	return form, nil
}

// List 列出所有儲存的表單資料
func (s *Storage) List() ([]*SavedForm, error) {
	rows, err := s.db.Query(`
		SELECT id, label, name, employee_id, start_date, end_date, leave_type, password, created_at, updated_at
		FROM saved_forms
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("查詢資料失敗: %w", err)
	}
	defer rows.Close()

	var forms []*SavedForm
	for rows.Next() {
		form := &SavedForm{}
		err := rows.Scan(
			&form.ID,
			&form.Label,
			&form.Name,
			&form.EmployeeID,
			&form.StartDate,
			&form.EndDate,
			&form.LeaveType,
			&form.Password,
			&form.CreatedAt,
			&form.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("讀取資料失敗: %w", err)
		}
		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("讀取資料失敗: %w", err)
	}

	return forms, nil
}

// Delete 刪除指定 ID 的表單資料
func (s *Storage) Delete(id int64) error {
	result, err := s.db.Exec("DELETE FROM saved_forms WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("刪除資料失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("確認刪除結果失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("找不到指定的資料")
	}

	return nil
}

// Update 更新表單資料
func (s *Storage) Update(form *SavedForm) error {
	now := time.Now()

	result, err := s.db.Exec(`
		UPDATE saved_forms
		SET label = ?, name = ?, employee_id = ?, start_date = ?, end_date = ?, leave_type = ?, password = ?, updated_at = ?
		WHERE id = ?
	`, form.Label, form.Name, form.EmployeeID, form.StartDate, form.EndDate, form.LeaveType, form.Password, now, form.ID)

	if err != nil {
		return fmt.Errorf("更新資料失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("確認更新結果失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("找不到指定的資料")
	}

	return nil
}

// Close 關閉資料庫連線
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
