### Task 9: 卡密服务 — 扣减（并发安全）

**Files:**
- Modify: `internal/service/key_service.go`

**Interfaces:**
- Produces: `ConsumeKey(rawKey, amount) (*ConsumeResult, int, error)` — 使用事务 + 乐观锁
- Consumes: `errcode.CodeKeyNotFound` 等 (Task 2)

- [ ] **Step 1: 在 key_service.go 末尾追加扣减方法**

```go
import (
	"CloudKey/internal/errcode"
	"CloudKey/internal/model"
	// ... 保持其他 import
)

type ConsumeResult struct {
	RemainingAmount int64           `json:"remaining_amount"`
	Status          model.KeyStatus `json:"status"`
	Exhausted       bool            `json:"exhausted"`
}

func (s *KeyService) ConsumeKey(rawKey string, amount int64) (*ConsumeResult, int, error) {
	key, err := s.FindByRawKey(rawKey)
	if err != nil {
		return nil, 0, err
	}
	if key == nil {
		return nil, errcode.CodeKeyNotFound, nil
	}
	if key.Status == model.KeyStatusDisabled {
		return nil, errcode.CodeKeyDisabled, nil
	}
	if key.Status == model.KeyStatusUsed {
		return nil, errcode.CodeKeyExhausted, nil
	}
	if !key.CanDeduct(amount) {
		if key.RemainingAmount <= 0 {
			return nil, errcode.CodeKeyExhausted, nil
		}
		return nil, errcode.CodeKeyInsufficient, nil
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Model(&model.Key{}).
		Where("id = ? AND version = ?", key.ID, key.Version).
		Updates(map[string]interface{}{
			"remaining_amount": gorm.Expr("remaining_amount - ?", amount),
			"version":          gorm.Expr("version + 1"),
			"status":           gorm.Expr("CASE WHEN remaining_amount - ? <= 0 THEN 'used' ELSE status END", amount),
			"used_at":          gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		tx.Rollback()
		return nil, 0, result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return s.ConsumeKey(rawKey, amount) // 乐观锁冲突，重试一次
	}

	var updatedKey model.Key
	if err := tx.Where("id = ?", key.ID).First(&updatedKey).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return &ConsumeResult{
		RemainingAmount: updatedKey.RemainingAmount,
		Status:          updatedKey.Status,
		Exhausted:       updatedKey.Status == model.KeyStatusUsed,
	}, 0, nil
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
git commit -m "feat(service): add ConsumeKey with transaction + optimistic locking"
```

---

