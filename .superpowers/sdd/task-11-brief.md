### Task 11: 使用记录 + 统计 + 管理员 + 服务账号 + 登录日志 + 配置服务

**Files:**
- Create: `internal/service/usage_log_service.go`
- Create: `internal/service/stats_service.go`
- Create: `internal/service/admin_service.go`
- Create: `internal/service/service_account_service.go`
- Create: `internal/service/login_log_service.go`
- Create: `internal/service/config_service.go`

**Interfaces:**
- Produces: `UsageLogService`, `StatsService`, `AdminService`, `ServiceAccountService`, `LoginLogService`, `ConfigService`

- [ ] **Step 1: 编写 usage_log_service.go**

```go
package service

import (
	"CloudKey/internal/model"
	"time"

	"gorm.io/gorm"
)

type UsageLogService struct {
	db *gorm.DB
}

func NewUsageLogService(db *gorm.DB) *UsageLogService {
	return &UsageLogService{db: db}
}

type RecordUsageParams struct {
	KeyID          uint64
	KeyAlias       string
	Amount         int64
	IP             string
	UserAgent      string
	RequestPath    string
	RequestParams  string
	ResponseStatus int
}

func (s *UsageLogService) Record(params RecordUsageParams) error {
	return s.db.Create(&model.UsageLog{
		KeyID:          params.KeyID,
		KeyAlias:       params.KeyAlias,
		Amount:         params.Amount,
		IP:             params.IP,
		UserAgent:      params.UserAgent,
		RequestPath:    params.RequestPath,
		RequestParams:  params.RequestParams,
		ResponseStatus: params.ResponseStatus,
		CreatedAt:      time.Now(),
	}).Error
}

type UsageLogQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	KeyAlias  string `form:"key_alias"`
	IP        string `form:"ip"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
}

func (s *UsageLogService) ListLogs(query UsageLogQuery) ([]model.UsageLog, int64, error) {
	var logs []model.UsageLog
	var total int64

	db := s.db.Model(&model.UsageLog{})
	if query.KeyAlias != "" {
		db = db.Where("key_alias = ?", query.KeyAlias)
	}
	if query.IP != "" {
		db = db.Where("ip = ?", query.IP)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (s *UsageLogService) ExportLogs(query UsageLogQuery) ([]model.UsageLog, error) {
	var logs []model.UsageLog
	db := s.db.Model(&model.UsageLog{})
	if query.KeyAlias != "" {
		db = db.Where("key_alias = ?", query.KeyAlias)
	}
	if query.IP != "" {
		db = db.Where("ip = ?", query.IP)
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}
	if err := db.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
```

- [ ] **Step 2: 编写 stats_service.go**

```go
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

	s.db.Model(&model.Key{}).Count(&ov.TotalKeys)

	var rows []struct {
		Status string
		Count  int64
	}
	s.db.Model(&model.Key{}).Select("status, COUNT(*) as count").Group("status").Scan(&rows)
	for _, r := range rows {
		ov.StatusCounts[r.Status] = r.Count
	}

	s.db.Model(&model.Key{}).Select("COALESCE(SUM(initial_amount), 0)").Scan(&ov.TotalInitial)
	s.db.Model(&model.Key{}).Select("COALESCE(SUM(remaining_amount), 0)").Scan(&ov.TotalRemain)

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
	s.db.Model(&model.UsageLog{}).
		Select("DATE_FORMAT(created_at, ?) as date, COUNT(*) as count", dateFormat).
		Where("created_at >= ?", startTime).
		Group("date").Order("date ASC").Scan(&points)

	return points, nil
}

type TopItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func (s *StatsService) GetTopKeys() ([]TopItem, error) {
	var items []TopItem
	s.db.Model(&model.UsageLog{}).
		Select("key_alias as name, COUNT(*) as count").
		Group("key_alias").Order("count DESC").Limit(10).Scan(&items)
	return items, nil
}

func (s *StatsService) GetTopIPs() ([]TopItem, error) {
	var items []TopItem
	s.db.Model(&model.UsageLog{}).
		Select("ip as name, COUNT(*) as count").
		Group("ip").Order("count DESC").Limit(10).Scan(&items)
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
	s.db.Model(&model.UsageLog{}).Where("created_at >= ?", todayStart).Count(&todayCalls)
	s.db.Model(&model.UsageLog{}).Where("created_at >= ?", weekStart).Count(&weekCalls)
	s.db.Model(&model.UsageLog{}).Where("created_at >= ?", monthStart).Count(&monthCalls)

	var recentLogs []model.UsageLog
	s.db.Order("created_at DESC").Limit(20).Find(&recentLogs)

	return &DashboardStats{
		Overview:   overview,
		TodayCalls: todayCalls,
		WeekCalls:  weekCalls,
		MonthCalls: monthCalls,
		RecentLogs: recentLogs,
	}, nil
}
```

- [ ] **Step 3: 编写 admin_service.go**

```go
package service

import (
	"CloudKey/internal/model"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	db          *gorm.DB
	jwtSecret   string
	jwtExpHours int
}

func NewAdminService(db *gorm.DB, jwtSecret string, jwtExpHours int) *AdminService {
	return &AdminService{db: db, jwtSecret: jwtSecret, jwtExpHours: jwtExpHours}
}

type LoginResult struct {
	RequireTOTP bool   `json:"require_totp"`
	AdminID     uint64 `json:"admin_id"`
}

func (s *AdminService) Login(username, password string) (*LoginResult, error) {
	var admin model.Admin
	if err := s.db.Where("username = ? AND is_active = ?", username, true).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, nil
	}

	return &LoginResult{RequireTOTP: admin.TotpSetup, AdminID: admin.ID}, nil
}

func (s *AdminService) VerifyTOTP(adminID uint64, code string) (string, error) {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return "", fmt.Errorf("admin not found: %w", err)
	}

	if !admin.TotpSetup {
		return "", fmt.Errorf("TOTP not set up")
	}

	if !totp.Validate(code, admin.TotpSecret) {
		return "", nil
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"username": admin.Username,
		"exp":      time.Now().Add(time.Duration(s.jwtExpHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}
	return tokenString, nil
}

func (s *AdminService) GenerateTOTPSecret(adminID uint64) (string, string, error) {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return "", "", err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CloudKey",
		AccountName: admin.Username,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate TOTP: %w", err)
	}

	if err := s.db.Model(&admin).Updates(map[string]interface{}{
		"totp_secret": key.Secret(),
		"totp_setup":  false,
	}).Error; err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *AdminService) ConfirmTOTPSetup(adminID uint64, code string) error {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return err
	}
	if !totp.Validate(code, admin.TotpSecret) {
		return fmt.Errorf("TOTP code invalid")
	}
	return s.db.Model(&admin).Update("totp_setup", true).Error
}

func (s *AdminService) ChangePassword(adminID uint64, oldPass, newPass string) error {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(oldPass)); err != nil {
		return fmt.Errorf("old password incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.db.Model(&admin).Update("password_hash", string(hash)).Error
}

func (s *AdminService) GetAdminProfile(adminID uint64) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminService) SeedAdmin(username, password string) error {
	var count int64
	s.db.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.db.Create(&model.Admin{
		Username:     username,
		PasswordHash: string(hash),
		TotpSetup:    false,
		IsActive:     true,
	}).Error
}
```

- [ ] **Step 4: 编写 service_account_service.go**

```go
package service

import (
	"CloudKey/internal/model"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"
)

type ServiceAccountService struct {
	db *gorm.DB
}

func NewServiceAccountService(db *gorm.DB) *ServiceAccountService {
	return &ServiceAccountService{db: db}
}

func hashServiceKey(key string) string {
	h := hmac.New(sha256.New, []byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *ServiceAccountService) GenerateServiceKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "svc-" + hex.EncodeToString(bytes), nil
}

func (s *ServiceAccountService) CreateServiceAccount(name string) (*model.ServiceAccount, string, error) {
	rawKey, err := s.GenerateServiceKey()
	if err != nil {
		return nil, "", err
	}

	account := model.ServiceAccount{Name: name, KeyHash: hashServiceKey(rawKey)}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, "", fmt.Errorf("create service account: %w", err)
	}
	return &account, rawKey, nil
}

func (s *ServiceAccountService) ValidateServiceKey(serviceKey string) (*model.ServiceAccount, error) {
	if serviceKey == "" {
		return nil, nil
	}

	keyHash := hashServiceKey(serviceKey)
	var account model.ServiceAccount
	if err := s.db.Where("key_hash = ? AND is_active = ?", keyHash, true).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (s *ServiceAccountService) ListServiceAccounts() ([]model.ServiceAccount, error) {
	var accounts []model.ServiceAccount
	if err := s.db.Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *ServiceAccountService) ToggleServiceAccount(id uint64, isActive bool) error {
	return s.db.Model(&model.ServiceAccount{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (s *ServiceAccountService) DeleteServiceAccount(id uint64) error {
	return s.db.Delete(&model.ServiceAccount{}, id).Error
}
```

- [ ] **Step 5: 编写 login_log_service.go**

```go
package service

import (
	"CloudKey/internal/model"

	"gorm.io/gorm"
)

type LoginLogService struct {
	db *gorm.DB
}

func NewLoginLogService(db *gorm.DB) *LoginLogService {
	return &LoginLogService{db: db}
}

func (s *LoginLogService) RecordLogin(adminID uint64, ip, userAgent string, success bool) error {
	status := model.LoginStatusFailed
	if success {
		status = model.LoginStatusSuccess
	}
	return s.db.Create(&model.LoginLog{
		AdminID: adminID, IP: ip, UserAgent: userAgent, Status: status,
	}).Error
}

func (s *LoginLogService) ListLoginLogs(page, pageSize int) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64

	s.db.Model(&model.LoginLog{}).Count(&total)
	offset := (page - 1) * pageSize
	s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	return logs, total, nil
}
```

- [ ] **Step 6: 编写 config_service.go**

```go
package service

import (
	"CloudKey/internal/model"

	"gorm.io/gorm"
)

type ConfigService struct {
	db *gorm.DB
}

func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

func (s *ConfigService) GetConfig(key string) (string, error) {
	var cfg model.SysConfig
	if err := s.db.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return cfg.Value, nil
}

func (s *ConfigService) GetAllConfigs() ([]model.SysConfig, error) {
	var configs []model.SysConfig
	if err := s.db.Order("id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *ConfigService) SetConfig(key, value, description string) error {
	var cfg model.SysConfig
	result := s.db.Where("`key` = ?", key).First(&cfg)
	if result.Error == gorm.ErrRecordNotFound {
		return s.db.Create(&model.SysConfig{Key: key, Value: value, Description: description}).Error
	}
	if result.Error != nil {
		return result.Error
	}
	updates := map[string]interface{}{"value": value}
	if description != "" {
		updates["description"] = description
	}
	return s.db.Model(&cfg).Updates(updates).Error
}

func (s *ConfigService) InitDefaultConfigs() error {
	defaults := []model.SysConfig{
		{Key: "key_prefix", Value: "sk-", Description: "卡密默认前缀"},
		{Key: "key_length", Value: "32", Description: "卡密随机部分长度"},
		{Key: "key_suffix_length", Value: "4", Description: "卡密后缀长度"},
		{Key: "record_request_params", Value: "false", Description: "是否记录请求参数"},
		{Key: "jwt_expire_hours", Value: "24", Description: "JWT 过期时间（小时）"},
	}

	for _, d := range defaults {
		var count int64
		s.db.Model(&model.SysConfig{}).Where("`key` = ?", d.Key).Count(&count)
		if count == 0 {
			if err := s.db.Create(&d).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
```

- [ ] **Step 7: 格式化并编译全部服务**

```bash
gofmt -w internal/service/
go build ./internal/service/
```

- [ ] **Step 8: 提交**

```bash
git add internal/service/
git commit -m "feat(service): add all service layer (usage_log, stats, admin, service_account, login_log, config)"
```

---

## 阶段三：服务层测试 (Task 12)

---

