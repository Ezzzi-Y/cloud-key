### Task 10: 卡密服务 — 管理操作

**Files:**
- Modify: `internal/service/key_service.go`

**Interfaces:**
- Produces: `ListKeys()`, `ListKeysByCreatedBy()`, `GetKeyDetail()`, `UpdateKey()`, `DisableKey()`, `EnableKey()`, `DeleteKey()`, `ExportKeys()`

- [ ] **Step 1: 在 key_service.go 末尾追加管理操作**

```go
type KeyListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Search   string `form:"search"`
}

func (s *KeyService) GetKeyDetail(id uint64) (*model.Key, error) {
	var key model.Key
	if err := s.db.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *KeyService) ListKeys(query KeyListQuery) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64

	db := s.db.Model(&model.Key{})
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Search != "" {
		db = db.Where("alias LIKE ? OR key_suffix LIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&keys).Error; err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

func (s *KeyService) ListKeysByCreatedBy(createdBy string, page, pageSize int) ([]model.Key, int64, error) {
	var keys []model.Key
	var total int64

	db := s.db.Model(&model.Key{}).Where("created_by = ?", createdBy)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys).Error; err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

type UpdateKeyRequest struct {
	Alias           *string `json:"alias"`
	RemainingAmount *int64  `json:"remaining_amount"`
}

func (s *KeyService) UpdateKey(id uint64, req UpdateKeyRequest) error {
	updates := map[string]interface{}{}
	if req.Alias != nil {
		updates["alias"] = *req.Alias
	}
	if req.RemainingAmount != nil {
		updates["remaining_amount"] = *req.RemainingAmount
		updates["status"] = gorm.Expr(
			"CASE WHEN ? > 0 AND status = 'used' THEN 'unused' ELSE status END",
			*req.RemainingAmount,
		)
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Model(&model.Key{}).Where("id = ?", id).Updates(updates).Error
}

func (s *KeyService) DisableKey(id uint64) error {
	return s.db.Model(&model.Key{}).Where("id = ?", id).Update("status", model.KeyStatusDisabled).Error
}

func (s *KeyService) EnableKey(id uint64) error {
	return s.db.Model(&model.Key{}).Where("id = ?", id).Update("status", model.KeyStatusUnused).Error
}

func (s *KeyService) DeleteKey(id uint64) error {
	return s.db.Delete(&model.Key{}, id).Error
}

func (s *KeyService) ExportKeys() ([]model.Key, error) {
	var keys []model.Key
	if err := s.db.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}
```

- [ ] **Step 2: 格式化并编译**

```bash
gofmt -w internal/service/key_service.go
go build ./internal/service/
```

- [ ] **Step 3: 提交**

```bash
git add internal/service/key_service.go
git commit -m "feat(service): add key management operations (list, update, disable, enable, delete, export)"
```

---

