package service

import (
	"CloudKey/internal/model"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StatsService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewStatsService(db *gorm.DB, rdb *redis.Client) *StatsService {
	return &StatsService{db: db, rdb: rdb}
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
		Select("COALESCE(SUM(remaining_amount), 0)").Scan(&ov.TotalRemain).Error; err != nil {
		return nil, err
	}

	return ov, nil
}

type TrendPoint struct {
	Date  string `json:"date"`
	Calls int64  `json:"calls"`
}

func (s *StatsService) GetTrends(period string, dateRange *DateRange, tenantID uint64) ([]TrendPoint, error) {
	// 自定义日期范围 → 走 MySQL（无法用固定桶覆盖任意区间）
	if dateRange != nil && dateRange.StartDate != "" {
		return s.getTrendsFromDB(period, dateRange, tenantID)
	}

	// today / week / month → 尝试 Redis
	if s.rdb != nil {
		points, err := s.getTrendsFromRedis(period, tenantID)
		if err == nil && len(points) > 0 {
			return points, nil
		}
		if err != nil {
			zap.L().Debug("trends Redis read failed, falling back to SQL", zap.Error(err))
		}
	}

	// Redis miss 或不可用 → MySQL 降级
	return s.getTrendsFromDB(period, dateRange, tenantID)
}

// getTrendsFromRedis 从 Redis Hash 读取趋势计数器。
func (s *StatsService) getTrendsFromRedis(period string, tenantID uint64) ([]TrendPoint, error) {
	ctx := context.Background()
	now := time.Now()

	switch period {
	case "today":
		key := trendHourlyKey(tenantID)
		vals, err := s.rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		todayPrefix := now.Format("2006-01-02")
		points := make([]TrendPoint, 0, 24)
		for h := 0; h < 24; h++ {
			field := fmt.Sprintf("%02d", h)
			calls := int64(0)
			if v, ok := vals[field]; ok {
				calls, _ = strconv.ParseInt(v, 10, 64)
			}
			points = append(points, TrendPoint{
				Date:  fmt.Sprintf("%s %s", todayPrefix, field),
				Calls: calls,
			})
		}
		return points, nil

	case "week":
		key := trendDailyKey(tenantID)
		vals, err := s.rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		points := make([]TrendPoint, 0, 7)
		for i := 6; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			field := d.Format("2006-01-02")
			calls := int64(0)
			if v, ok := vals[field]; ok {
				calls, _ = strconv.ParseInt(v, 10, 64)
			}
			points = append(points, TrendPoint{Date: field, Calls: calls})
		}
		return points, nil

	case "month":
		key := trendDailyKey(tenantID)
		vals, err := s.rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		points := make([]TrendPoint, 0, 30)
		for i := 29; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			field := d.Format("2006-01-02")
			calls := int64(0)
			if v, ok := vals[field]; ok {
				calls, _ = strconv.ParseInt(v, 10, 64)
			}
			points = append(points, TrendPoint{Date: field, Calls: calls})
		}
		return points, nil
	}

	return nil, nil
}

// getTrendsFromDB 从 MySQL usage_logs 表查询趋势（支持日期过滤和 month 视图）。
func (s *StatsService) getTrendsFromDB(period string, dateRange *DateRange, tenantID uint64) ([]TrendPoint, error) {
	var dateFormat string
	var startTime time.Time
	var points []TrendPoint
	now := time.Now()

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
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as calls", dateFormat).
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

// trendHourlyKey returns the Redis Hash key for hourly trend counters.
func trendHourlyKey(tenantID uint64) string {
	return fmt.Sprintf("ck:trend:hourly:%d", tenantID)
}

// trendDailyKey returns the Redis Hash key for daily trend counters.
func trendDailyKey(tenantID uint64) string {
	return fmt.Sprintf("ck:trend:daily:%d", tenantID)
}

// endOfDay returns the end of the given day (23:59:59) for Redis TTL.
func endOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, t.Location())
}

type TopItem struct {
	KeyAlias string `json:"key_alias"`
	Count    int64  `json:"count"`
}

func (s *StatsService) GetTopKeys(dateRange *DateRange, tenantID uint64) ([]TopItem, error) {
	// 无日期过滤时优先走 Redis 缓存
	if dateRange == nil {
		items, err := s.getTopKeysFromRedis(tenantID)
		if err == nil && len(items) > 0 {
			return items, nil
		}
		if err != nil {
			zap.L().Debug("top_keys Redis read failed, falling back to SQL", zap.Error(err))
		}
	}

	// SQL fallback（支持日期过滤 或 Redis miss）
	return s.getTopKeysFromDB(dateRange, tenantID)
}

// getTopKeysFromRedis 从 Redis ZSET 读取 top 10 热门卡密，批量解析 alias。
func (s *StatsService) getTopKeysFromRedis(tenantID uint64) ([]TopItem, error) {
	ctx := context.Background()
	zsetKey := topKeysZSetKey(tenantID)

	// ZREVRANGE key 0 9 WITHSCORES — O(log N + 10)
	zs, err := s.rdb.ZRevRangeWithScores(ctx, zsetKey, 0, 9).Result()
	if err != nil {
		return nil, err
	}
	if len(zs) == 0 {
		return nil, nil
	}

	// 收集 key ID，批量查询 alias
	ids := make([]uint64, 0, len(zs))
	for _, z := range zs {
		id, err := strconv.ParseUint(z.Member.(string), 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	type keyInfo struct {
		ID    uint64
		Alias string
	}
	var keys []keyInfo
	if err := s.db.Model(&model.Key{}).
		Where("id IN ?", ids).
		Select("id, alias").
		Scan(&keys).Error; err != nil {
		return nil, err
	}

	aliasMap := make(map[uint64]string, len(keys))
	for _, k := range keys {
		aliasMap[k.ID] = k.Alias
	}

	items := make([]TopItem, 0, len(zs))
	for _, z := range zs {
		id, _ := strconv.ParseUint(z.Member.(string), 10, 64)
		alias, ok := aliasMap[id]
		if !ok {
			continue // key 已被删除，跳过
		}
		items = append(items, TopItem{
			KeyAlias: alias,
			Count:    int64(z.Score),
		})
	}

	return items, nil
}

// getTopKeysFromDB 从 usage_logs 表查询 top 10（支持日期过滤）。
func (s *StatsService) getTopKeysFromDB(dateRange *DateRange, tenantID uint64) ([]TopItem, error) {
	items := make([]TopItem, 0)
	db := applyDateFilter(s.db.Model(&model.UsageLog{}), dateRange).Where("tenant_id = ?", tenantID)
	if err := db.
		Select("key_alias, COUNT(*) as count").
		Group("key_alias").Order("count DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

type TopAmountItem struct {
	KeyAlias string `json:"key_alias"`
	Amount   int64  `json:"total_amount"`
}

func (s *StatsService) GetTopAmount(dateRange *DateRange, tenantID uint64) ([]TopAmountItem, error) {
	// 无日期过滤时优先走 Redis 缓存
	if dateRange == nil {
		items, err := s.getTopAmountFromRedis(tenantID)
		if err == nil && len(items) > 0 {
			return items, nil
		}
		if err != nil {
			zap.L().Debug("top_amount Redis read failed, falling back to SQL", zap.Error(err))
		}
	}

	// SQL fallback（支持日期过滤 或 Redis miss）
	return s.getTopAmountFromDB(dateRange, tenantID)
}

// getTopAmountFromRedis 从 Redis ZSET 读取 top 10 额度消耗卡密，批量解析 alias。
func (s *StatsService) getTopAmountFromRedis(tenantID uint64) ([]TopAmountItem, error) {
	ctx := context.Background()
	zsetKey := topAmountZSetKey(tenantID)

	// ZREVRANGE key 0 9 WITHSCORES — O(log N + 10)
	zs, err := s.rdb.ZRevRangeWithScores(ctx, zsetKey, 0, 9).Result()
	if err != nil {
		return nil, err
	}
	if len(zs) == 0 {
		return nil, nil
	}

	// 收集 key ID，批量查询 alias
	ids := make([]uint64, 0, len(zs))
	for _, z := range zs {
		id, err := strconv.ParseUint(z.Member.(string), 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	type keyInfo struct {
		ID    uint64
		Alias string
	}
	var keys []keyInfo
	if err := s.db.Model(&model.Key{}).
		Where("id IN ?", ids).
		Select("id, alias").
		Scan(&keys).Error; err != nil {
		return nil, err
	}

	aliasMap := make(map[uint64]string, len(keys))
	for _, k := range keys {
		aliasMap[k.ID] = k.Alias
	}

	items := make([]TopAmountItem, 0, len(zs))
	for _, z := range zs {
		id, _ := strconv.ParseUint(z.Member.(string), 10, 64)
		alias, ok := aliasMap[id]
		if !ok {
			continue // key 已被删除，跳过
		}
		items = append(items, TopAmountItem{
			KeyAlias: alias,
			Amount:   int64(z.Score),
		})
	}

	return items, nil
}

// getTopAmountFromDB 从 usage_logs 表查询 top 10 额度消耗（支持日期过滤）。
func (s *StatsService) getTopAmountFromDB(dateRange *DateRange, tenantID uint64) ([]TopAmountItem, error) {
	items := make([]TopAmountItem, 0)
	db := applyDateFilter(s.db.Model(&model.UsageLog{}), dateRange).Where("tenant_id = ?", tenantID)
	if err := db.
		Select("key_alias, SUM(amount) as total_amount").
		Group("key_alias").Order("total_amount DESC").Limit(10).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

type DashboardStats struct {
	KeyCount        int64            `json:"key_count"`
	StatusBreakdown map[string]int64 `json:"key_status_breakdown"`
	TodayCalls      int64            `json:"today_calls"`
	WeekCalls       int64            `json:"week_calls"`
	MonthCalls      int64            `json:"month_calls"`
	CanRefresh      bool             `json:"can_refresh"`
	NextRefreshAt   *time.Time       `json:"next_refresh_at"`
	RecentLogs      []model.UsageLog `json:"recent_logs"`
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
		KeyCount:        overview.KeyCount,
		StatusBreakdown: overview.StatusCounts,
		TodayCalls:      todayCalls,
		WeekCalls:       weekCalls,
		MonthCalls:      monthCalls,
		CanRefresh:      s.canRefreshTop(tenantID),
		NextRefreshAt:   s.nextRefreshAt(tenantID),
		RecentLogs:      recentLogs,
	}, nil
}

func (s *StatsService) canRefreshTop(tenantID uint64) bool {
	if s.rdb == nil {
		return true
	}
	ctx := context.Background()
	exists, err := s.rdb.Exists(ctx, refreshTopLockKey(tenantID)).Result()
	if err != nil {
		return true // 出错时允许操作
	}
	return exists == 0
}

func (s *StatsService) nextRefreshAt(tenantID uint64) *time.Time {
	if s.rdb == nil {
		return nil
	}
	ctx := context.Background()
	ttl, err := s.rdb.TTL(ctx, refreshTopLockKey(tenantID)).Result()
	if err != nil || ttl <= 0 {
		return nil
	}
	next := time.Now().Add(ttl)
	return &next
}
