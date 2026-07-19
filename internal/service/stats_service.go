package service

import (
	"CloudKey/internal/model"
	"time"

	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

type KeyOverview struct {
	TotalKeys    int64            `json:"total_keys"`
	StatusCounts map[string]int64 `json:"status_counts"`
	TotalInitial int64            `json:"total_initial"`
	TotalRemain  int64            `json:"total_remaining"`
}

func (s *StatsService) GetKeyOverview() (*KeyOverview, error) {
	ov := &KeyOverview{StatusCounts: make(map[string]int64)}

	if err := s.db.Model(&model.Key{}).Count(&ov.TotalKeys).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		Status string
		Count  int64
	}
	if err := s.db.Model(&model.Key{}).Select("status, COUNT(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		ov.StatusCounts[r.Status] = r.Count
	}

	if err := s.db.Model(&model.Key{}).Select("COALESCE(SUM(initial_amount), 0)").Scan(&ov.TotalInitial).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Key{}).Select("COALESCE(SUM(remaining_amount), 0)").Scan(&ov.TotalRemain).Error; err != nil {
		return nil, err
	}

	return ov, nil
}

type TrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func (s *StatsService) GetTrends(period string) ([]TrendPoint, error) {
	var dateFormat string
	var startTime time.Time
	now := time.Now()

	switch period {
	case "week":
		dateFormat = "%Y-%m-%d"
		startTime = now.AddDate(0, 0, -7)
	case "month":
		dateFormat = "%Y-%m-%d"
		startTime = now.AddDate(0, -1, 0)
	default:
		dateFormat = "%Y-%m-%d %H"
		startTime = now.AddDate(0, 0, -1)
	}

	var points []TrendPoint
	if err := s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as count", dateFormat).
		Where("created_at >= ?", startTime).
		Group("date").Order("date ASC").Scan(&points).Error; err != nil {
		return nil, err
	}

	return points, nil
}

type TopItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func (s *StatsService) GetTopKeys() ([]TopItem, error) {
	var items []TopItem
	if err := s.db.Model(&model.UsageLog{}).
		Select("key_alias as name, COUNT(*) as count").
		Group("key_alias").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *StatsService) GetTopIPs() ([]TopItem, error) {
	var items []TopItem
	if err := s.db.Model(&model.UsageLog{}).
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

func (s *StatsService) GetDashboard() (*DashboardStats, error) {
	overview, err := s.GetKeyOverview()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, -1, 0)

	var todayCalls, weekCalls, monthCalls int64
	if err := s.db.Model(&model.UsageLog{}).Where("created_at >= ?", todayStart).Count(&todayCalls).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.UsageLog{}).Where("created_at >= ?", weekStart).Count(&weekCalls).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.UsageLog{}).Where("created_at >= ?", monthStart).Count(&monthCalls).Error; err != nil {
		return nil, err
	}

	var recentLogs []model.UsageLog
	if err := s.db.Order("created_at DESC").Limit(20).Find(&recentLogs).Error; err != nil {
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
