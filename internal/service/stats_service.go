package service

import (
	"CloudKey/internal/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

// DateRange holds optional start/end date filters.
// Empty strings mean no filter on that boundary.
type DateRange struct {
	StartDate string
	EndDate   string
}

func applyDateFilter(db *gorm.DB, dateRange *DateRange) *gorm.DB {
	if dateRange == nil {
		return db
	}
	if dateRange.StartDate != "" {
		db = db.Where("created_at >= ?", dateRange.StartDate)
	}
	if dateRange.EndDate != "" {
		db = db.Where("created_at <= ?", dateRange.EndDate)
	}
	return db
}

type KeyOverview struct {
	KeyCount     int64            `json:"key_count"`
	StatusCounts map[string]int64 `json:"status_counts"`
	TotalInitial int64            `json:"total_initial"`
	TotalRemain  int64            `json:"total_remaining"`
}

func (s *StatsService) GetKeyOverview(dateRange *DateRange, tenantID uint64) (*KeyOverview, error) {
	ov := &KeyOverview{StatusCounts: make(map[string]int64)}

	keyDB := applyDateFilter(s.db.Model(&model.Key{}), dateRange).Where("tenant_id = ?", tenantID)
	if err := keyDB.Count(&ov.KeyCount).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		Status string
		Count  int64
	}
	if err := applyDateFilter(s.db.Model(&model.Key{}), dateRange).
		Where("tenant_id = ?", tenantID).
		Select("status, COUNT(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		ov.StatusCounts[r.Status] = r.Count
	}

	if err := applyDateFilter(s.db.Model(&model.Key{}), dateRange).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(initial_amount), 0)").Scan(&ov.TotalInitial).Error; err != nil {
		return nil, err
	}
	if err := applyDateFilter(s.db.Model(&model.Key{}), dateRange).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(remaining_amount), 0)").Scan(&ov.TotalRemain).Error; err != nil {
		return nil, err
	}

	return ov, nil
}

type TrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (s *StatsService) GetTrends(period string, dateRange *DateRange, tenantID uint64) ([]TrendPoint, error) {
	var dateFormat string
	var startTime time.Time
	var points []TrendPoint
	now := time.Now()

	// If explicit date range is provided, use it instead of period
	if dateRange != nil && dateRange.StartDate != "" {
		dateFormat = "%Y-%m-%d"
		parsed, err := time.Parse("2006-01-02", dateRange.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		startTime = parsed
	} else {
		switch period {
		case "week":
			dateFormat = "%Y-%m-%d"
			startTime = now.AddDate(0, 0, -7)
		case "month":
			dateFormat = "%Y-%m-%d"
			startTime = now.AddDate(0, -1, 0)
		default: // "today"
			dateFormat = "%Y-%m-%d %H"
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		}
	}

	db := s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as count", dateFormat).
		Where("created_at >= ?", startTime).
		Where("tenant_id = ?", tenantID)
	db = applyDateFilter(db, dateRange)

	if err := db.Group("date").Order("date ASC").Scan(&points).Error; err != nil {
		return nil, err
	}

	if points == nil {
		points = make([]TrendPoint, 0)
	}

	return points, nil
}

type TopItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func (s *StatsService) GetTopKeys(dateRange *DateRange, tenantID uint64) ([]TopItem, error) {
	items := make([]TopItem, 0)
	db := applyDateFilter(s.db.Model(&model.UsageLog{}), dateRange).Where("tenant_id = ?", tenantID)
	if err := db.
		Select("key_alias as name, COUNT(*) as count").
		Group("key_alias").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *StatsService) GetTopIPs(dateRange *DateRange, tenantID uint64) ([]TopItem, error) {
	items := make([]TopItem, 0)
	db := applyDateFilter(s.db.Model(&model.UsageLog{}), dateRange).Where("tenant_id = ?", tenantID)
	if err := db.
		Select("ip as name, COUNT(*) as count").
		Group("ip").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

type DashboardStats struct {
	Overview   *KeyOverview     `json:"overview"`
	TodayCalls int64            `json:"today_calls"`
	WeekCalls  int64            `json:"week_calls"`
	MonthCalls int64            `json:"month_calls"`
	RecentLogs []model.UsageLog `json:"recent_logs"`
}

func (s *StatsService) GetDashboard(tenantID uint64) (*DashboardStats, error) {
	overview, err := s.GetKeyOverview(nil, tenantID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, -1, 0)

	var todayCalls, weekCalls, monthCalls int64
	if err := s.db.Model(&model.UsageLog{}).Where("created_at >= ?", todayStart).Where("tenant_id = ?", tenantID).Count(&todayCalls).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.UsageLog{}).Where("created_at >= ?", weekStart).Where("tenant_id = ?", tenantID).Count(&weekCalls).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.UsageLog{}).Where("created_at >= ?", monthStart).Where("tenant_id = ?", tenantID).Count(&monthCalls).Error; err != nil {
		return nil, err
	}

	var recentLogs []model.UsageLog
	if err := s.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Limit(20).Find(&recentLogs).Error; err != nil {
		return nil, err
	}

	return &DashboardStats{
		Overview:   overview,
		TodayCalls: todayCalls,
		WeekCalls:  weekCalls,
		MonthCalls: monthCalls,
		RecentLogs: recentLogs,
	}, nil
}
