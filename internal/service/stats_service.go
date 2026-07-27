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
	Date   string `json:"date"`
	Calls  int64  `json:"calls"`
	Amount int64  `json:"amount"`
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
		callsKey := trendHourlyKey(tenantID)
		amountKey := trendHourlyAmountKey(tenantID)
		callsVals, err := s.rdb.HGetAll(ctx, callsKey).Result()
		if err != nil {
			return nil, err
		}
		amountVals, err := s.rdb.HGetAll(ctx, amountKey).Result()
		if err != nil {
			return nil, err
		}
		todayPrefix := now.Format("2006-01-02")
		points := make([]TrendPoint, 0, 24)
		for h := 0; h < 24; h++ {
			field := fmt.Sprintf("%02d", h)
			calls := int64(0)
			if v, ok := callsVals[field]; ok {
				calls, _ = strconv.ParseInt(v, 10, 64)
			}
			amount := int64(0)
			if v, ok := amountVals[field]; ok {
				amount, _ = strconv.ParseInt(v, 10, 64)
			}
			points = append(points, TrendPoint{
				Date:   fmt.Sprintf("%s %s", todayPrefix, field),
				Calls:  calls,
				Amount: amount,
			})
		}
		return points, nil

	case "week":
		callsKey := trendDailyKey(tenantID)
		amountKey := trendDailyAmountKey(tenantID)
		callsVals, err := s.rdb.HGetAll(ctx, callsKey).Result()
		if err != nil {
			return nil, err
		}
		amountVals, err := s.rdb.HGetAll(ctx, amountKey).Result()
		if err != nil {
			return nil, err
		}
		points := make([]TrendPoint, 0, 7)
		for i := 6; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			field := d.Format("2006-01-02")
			calls := int64(0)
			if v, ok := callsVals[field]; ok {
				calls, _ = strconv.ParseInt(v, 10, 64)
			}
			amount := int64(0)
			if v, ok := amountVals[field]; ok {
				amount, _ = strconv.ParseInt(v, 10, 64)
			}
			points = append(points, TrendPoint{Date: field, Calls: calls, Amount: amount})
		}
		return points, nil

	case "month":
		callsKey := trendDailyKey(tenantID)
		amountKey := trendDailyAmountKey(tenantID)
		callsVals, err := s.rdb.HGetAll(ctx, callsKey).Result()
		if err != nil {
			return nil, err
		}
		amountVals, err := s.rdb.HGetAll(ctx, amountKey).Result()
		if err != nil {
			return nil, err
		}
		points := make([]TrendPoint, 0, 30)
		for i := 29; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			field := d.Format("2006-01-02")
			calls := int64(0)
			if v, ok := callsVals[field]; ok {
				calls, _ = strconv.ParseInt(v, 10, 64)
			}
			amount := int64(0)
			if v, ok := amountVals[field]; ok {
				amount, _ = strconv.ParseInt(v, 10, 64)
			}
			points = append(points, TrendPoint{Date: field, Calls: calls, Amount: amount})
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
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as calls, COALESCE(SUM(amount), 0) as amount", dateFormat).
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

// trendHourlyAmountKey returns the Redis Hash key for hourly trend amount counters.
func trendHourlyAmountKey(tenantID uint64) string {
	return fmt.Sprintf("ck:trend:hourly_amount:%d", tenantID)
}

// trendDailyAmountKey returns the Redis Hash key for daily trend amount counters.
func trendDailyAmountKey(tenantID uint64) string {
	return fmt.Sprintf("ck:trend:daily_amount:%d", tenantID)
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

// keyUsageCallsKey 返回单 Key 每日调用次数 ZSET key。
func keyUsageCallsKey(tenantID, keyID uint64) string {
	return fmt.Sprintf("key_usage:calls:%d:%d", tenantID, keyID)
}

// keyUsageAmountKey 返回单 Key 每日额度消耗 ZSET key。
func keyUsageAmountKey(tenantID, keyID uint64) string {
	return fmt.Sprintf("key_usage:amount:%d:%d", tenantID, keyID)
}

// keyUsageRefreshLockKey 返回限制刷新频率的 Redis key。
func keyUsageRefreshLockKey(tenantID, keyID uint64) string {
	return fmt.Sprintf("key_usage_refresh_lock:%d:%d", tenantID, keyID)
}

// KeyUsagePoint 单日使用数据点。
type KeyUsagePoint struct {
	Date   string `json:"date"`
	Calls  int64  `json:"calls"`
	Amount int64  `json:"amount"`
}

// KeyUsageStats 单个 Key 的使用情况统计。
type KeyUsageStats struct {
	KeyAlias      string          `json:"key_alias"`
	KeySuffix     string          `json:"key_suffix"`
	Points        []KeyUsagePoint `json:"points"`
	CanRefresh    bool            `json:"can_refresh"`
	NextRefreshAt *time.Time      `json:"next_refresh_at"`
}

// GetKeyUsage 查询单个 Key 近 30 天的使用情况（调用次数 + 额度消耗）。
// 优先从 Redis ZSET 读取，miss 时降级到 MySQL。
func (s *StatsService) GetKeyUsage(tenantID, keyID uint64) (*KeyUsageStats, error) {
	// 查询 key 基本信息
	var keyInfo struct {
		Alias    string
		KeySuffix string
	}
	if err := s.db.Model(&model.Key{}).
		Where("id = ? AND tenant_id = ?", keyID, tenantID).
		Select("alias, key_suffix").
		Scan(&keyInfo).Error; err != nil {
		return nil, err
	}

	stats := &KeyUsageStats{
		KeyAlias:   keyInfo.Alias,
		KeySuffix:  keyInfo.KeySuffix,
		CanRefresh: s.canRefreshKeyUsage(tenantID, keyID),
		NextRefreshAt: s.keyUsageNextRefreshAt(tenantID, keyID),
	}

	// 尝试从 Redis 读取
	if s.rdb != nil {
		points, err := s.getKeyUsageFromRedis(tenantID, keyID)
		if err == nil && len(points) > 0 {
			stats.Points = points
			return stats, nil
		}
		if err != nil {
			zap.L().Debug("key_usage Redis read failed, falling back to SQL", zap.Error(err))
		}
	}

	// Redis miss → MySQL 降级
	points, err := s.getKeyUsageFromDB(tenantID, keyID)
	if err != nil {
		return nil, err
	}
	stats.Points = points
	return stats, nil
}

// getKeyUsageFromRedis 从 Redis ZSET 读取单 Key 近 30 天使用数据。
func (s *StatsService) getKeyUsageFromRedis(tenantID, keyID uint64) ([]KeyUsagePoint, error) {
	ctx := context.Background()
	now := time.Now()
	startDate := now.AddDate(0, 0, -29).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	callsKey := keyUsageCallsKey(tenantID, keyID)
	amountKey := keyUsageAmountKey(tenantID, keyID)

	callsScores, err := s.rdb.ZRangeByScoreWithScores(ctx, callsKey, &redis.ZRangeBy{
		Min: startDate, Max: endDate,
	}).Result()
	if err != nil {
		return nil, err
	}
	amountScores, err := s.rdb.ZRangeByScoreWithScores(ctx, amountKey, &redis.ZRangeBy{
		Min: startDate, Max: endDate,
	}).Result()
	if err != nil {
		return nil, err
	}

	// 如果完全没有数据，返回 nil 触发 fallback
	if len(callsScores) == 0 && len(amountScores) == 0 {
		return nil, nil
	}

	// 构建 map
	callsMap := make(map[string]int64, len(callsScores))
	for _, z := range callsScores {
		callsMap[z.Member.(string)] = int64(z.Score)
	}
	amountMap := make(map[string]int64, len(amountScores))
	for _, z := range amountScores {
		amountMap[z.Member.(string)] = int64(z.Score)
	}

	// 填充 30 天连续数据
	points := make([]KeyUsagePoint, 0, 30)
	for i := 29; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		date := d.Format("2006-01-02")
		points = append(points, KeyUsagePoint{
			Date:   date,
			Calls:  callsMap[date],
			Amount: amountMap[date],
		})
	}
	return points, nil
}

// getKeyUsageFromDB 从 MySQL usage_logs 查询单 Key 近 30 天使用数据。
func (s *StatsService) getKeyUsageFromDB(tenantID, keyID uint64) ([]KeyUsagePoint, error) {
	now := time.Now()
	startTime := now.AddDate(0, 0, -29)
	startDate := startTime.Format("2006-01-02")

	type dbRow struct {
		Date   string
		Calls  int64
		Amount int64
	}
	var rows []dbRow
	if err := s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, COUNT(*) as calls, COALESCE(SUM(amount), 0) as amount").
		Where("tenant_id = ? AND key_id = ? AND created_at >= ?", tenantID, keyID, startDate).
		Group("date").Order("date ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	rowMap := make(map[string]dbRow, len(rows))
	for _, r := range rows {
		rowMap[r.Date] = r
	}

	// 填充 30 天连续数据
	points := make([]KeyUsagePoint, 0, 30)
	for i := 29; i >= 0; i-- {
		d := now.AddDate(0, 0, -i)
		date := d.Format("2006-01-02")
		p := KeyUsagePoint{Date: date}
		if r, ok := rowMap[date]; ok {
			p.Calls = r.Calls
			p.Amount = r.Amount
		}
		points = append(points, p)
	}
	return points, nil
}

// canRefreshKeyUsage 检查该 Key 今天是否还能刷新使用统计。
func (s *StatsService) canRefreshKeyUsage(tenantID, keyID uint64) bool {
	if s.rdb == nil {
		return true
	}
	ctx := context.Background()
	exists, err := s.rdb.Exists(ctx, keyUsageRefreshLockKey(tenantID, keyID)).Result()
	if err != nil {
		return true
	}
	return exists == 0
}

// keyUsageNextRefreshAt 返回下次可刷新的时间。
func (s *StatsService) keyUsageNextRefreshAt(tenantID, keyID uint64) *time.Time {
	if s.rdb == nil {
		return nil
	}
	ctx := context.Background()
	ttl, err := s.rdb.TTL(ctx, keyUsageRefreshLockKey(tenantID, keyID)).Result()
	if err != nil || ttl <= 0 {
		return nil
	}
	next := time.Now().Add(ttl)
	return &next
}

// RefreshKeyUsage 从 MySQL 重建单 Key 近 30 天使用统计到 Redis ZSET。
// 限频：同一个 Key 每 24 小时只能刷新一次。
func (s *StatsService) RefreshKeyUsage(tenantID, keyID uint64) error {
	if s.rdb == nil {
		return fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	lockKey := keyUsageRefreshLockKey(tenantID, keyID)

	// 尝试获取锁（SETNX + 24h TTL）
	ok, err := s.rdb.SetNX(ctx, lockKey, 1, 24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("setnx lock: %w", err)
	}
	if !ok {
		return fmt.Errorf("already refreshed")
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -29).Format("2006-01-02")

	// 查询调用次数
	type dayCount struct {
		Date  string
		Count int64
	}
	var callRows []dayCount
	if err := s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, COUNT(*) as count").
		Where("tenant_id = ? AND key_id = ? AND created_at >= ?", tenantID, keyID, startDate).
		Group("date").Scan(&callRows).Error; err != nil {
		return fmt.Errorf("query calls: %w", err)
	}

	// 查询额度消耗
	type dayAmount struct {
		Date   string
		Amount int64
	}
	var amountRows []dayAmount
	if err := s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, COALESCE(SUM(amount), 0) as amount").
		Where("tenant_id = ? AND key_id = ? AND created_at >= ?", tenantID, keyID, startDate).
		Group("date").Scan(&amountRows).Error; err != nil {
		return fmt.Errorf("query amount: %w", err)
	}

	// 构建 ZSET members
	callMembers := make([]redis.Z, 0, len(callRows))
	for _, r := range callRows {
		callMembers = append(callMembers, redis.Z{Score: float64(r.Count), Member: r.Date})
	}
	amountMembers := make([]redis.Z, 0, len(amountRows))
	for _, r := range amountRows {
		amountMembers = append(amountMembers, redis.Z{Score: float64(r.Amount), Member: r.Date})
	}

	callsKey := keyUsageCallsKey(tenantID, keyID)
	amountKey := keyUsageAmountKey(tenantID, keyID)

	// 原子替换：先删除旧数据，再写入新数据
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, callsKey)
	pipe.Del(ctx, amountKey)
	if len(callMembers) > 0 {
		pipe.ZAdd(ctx, callsKey, callMembers...)
	}
	if len(amountMembers) > 0 {
		pipe.ZAdd(ctx, amountKey, amountMembers...)
	}
	pipe.Expire(ctx, callsKey, 31*24*time.Hour)
	pipe.Expire(ctx, amountKey, 31*24*time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec: %w", err)
	}

	zap.L().Info("key usage 统计刷新完成",
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("key_id", keyID),
		zap.Int("call_days", len(callMembers)),
		zap.Int("amount_days", len(amountMembers)))
	return nil
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
