package service

import (
	"CloudKey/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MQWorker 消费者 Worker，处理 consume 和 adjust 队列的消息
type MQWorker struct {
	mq     *MQService
	db     *gorm.DB
	rdb    *redis.Client
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewMQWorker 创建消费者 Worker
func NewMQWorker(mq *MQService, db *gorm.DB, rdb *redis.Client) *MQWorker {
	return &MQWorker{
		mq:  mq,
		db:  db,
		rdb: rdb,
	}
}

// Start 启动两个 goroutine 分别消费 consume 和 adjust 队列
func (w *MQWorker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	w.wg.Add(2)
	go w.consumeLoop(ctx, mqConsumeQueue, w.handleConsume)
	go w.consumeLoop(ctx, mqAdjustQueue, w.handleAdjust)

	zap.L().Info("MQ Worker 已启动",
		zap.String("consume_queue", mqConsumeQueue),
		zap.String("adjust_queue", mqAdjustQueue),
	)
}

// Stop 优雅停止 Worker
func (w *MQWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	zap.L().Info("MQ Worker 已停止")
}

// consumeLoop 通用消费循环
func (w *MQWorker) consumeLoop(ctx context.Context, queue string, handler func(amqp.Delivery) error) {
	defer w.wg.Done()

	ch, err := w.mq.conn.Channel()
	if err != nil {
		zap.L().Error("打开消费 channel 失败", zap.String("queue", queue), zap.Error(err))
		return
	}
	defer ch.Close()

	if err := ch.Qos(mqPrefetch, 0, false); err != nil {
		zap.L().Error("设置 prefetch 失败", zap.String("queue", queue), zap.Error(err))
		return
	}

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		zap.L().Error("注册消费者失败", zap.String("queue", queue), zap.Error(err))
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			if err := handler(d); err != nil {
				zap.L().Error("处理消息失败",
					zap.String("queue", queue),
					zap.String("message_id", d.MessageId),
					zap.Error(err),
				)
				// 重试次数判断
				retryCount := getRetryCount(d.Headers)
				if retryCount < mqMaxRetries {
					// requeue 重试
					_ = d.Nack(false, true)
				} else {
					// 超过重试次数，发到 DLQ（通过 Nack + 不 requeue，由死信路由处理）
					zap.L().Warn("消息超过最大重试次数，发送到 DLQ",
						zap.String("message_id", d.MessageId),
						zap.Int("retry_count", retryCount),
					)
					_ = d.Nack(false, false)
				}
			} else {
				_ = d.Ack(false)
			}
		}
	}
}

// getRetryCount 从消息 headers 中获取重试次数
func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if count, ok := headers["x-death"]; ok {
		if deaths, ok := count.(amqp.Table); ok {
			if rc, ok := deaths["count"]; ok {
				if n, ok := rc.(int64); ok {
					return int(n)
				}
				if n, ok := rc.(int32); ok {
					return int(n)
				}
			}
		}
	}
	// 也检查自定义 header
	if rc, ok := headers["x-retry-count"]; ok {
		if n, ok := rc.(int64); ok {
			return int(n)
		}
		if n, ok := rc.(int32); ok {
			return int(n)
		}
	}
	return 0
}

// handleConsume 处理消费扣减事件
func (w *MQWorker) handleConsume(d amqp.Delivery) error {
	var event ConsumeEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		zap.L().Error("反序列化 ConsumeEvent 失败", zap.Error(err))
		return nil // 格式错误不重试，直接丢弃
	}

	tx := w.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx: %w", tx.Error)
	}

	// 更新 key 余额（增量扣减）和状态
	result := tx.Model(&model.Key{}).Where("id = ?", event.KeyID).
		Updates(map[string]interface{}{
			"remaining_amount": gorm.Expr("remaining_amount - ?", event.Amount),
			"status":           event.StatusAfter,
			"used_at":          time.UnixMilli(event.UsedAt),
		})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("update key: %w", result.Error)
	}

	// 写入消费日志
	log := model.UsageLog{
		TenantID:       event.TenantID,
		KeyID:          event.KeyID,
		KeyAlias:       event.KeyAlias,
		Amount:         event.Amount,
		IP:             event.IP,
		UserAgent:      event.UserAgent,
		RequestPath:    event.Path,
		ResponseStatus: http.StatusOK, // 消费成功才发 MQ
		CreatedAt:      time.UnixMilli(event.Timestamp),
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("insert usage_log: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// 异步更新趋势计数器（写入 Redis，失败仅记日志）
	if w.rdb != nil {
		now := time.UnixMilli(event.Timestamp)
		ctx := context.Background()
		hourKey := trendHourlyKey(event.TenantID)
		dayKey := trendDailyKey(event.TenantID)
		if err := w.rdb.HIncrBy(ctx, hourKey, now.Format("15"), 1).Err(); err != nil {
			zap.L().Debug("trend hourly HINCRBY failed", zap.Error(err))
		} else {
			w.rdb.ExpireAt(ctx, hourKey, endOfDay(now))
		}
		if err := w.rdb.HIncrBy(ctx, dayKey, now.Format("2006-01-02"), 1).Err(); err != nil {
			zap.L().Debug("trend daily HINCRBY failed", zap.Error(err))
		} else {
			w.rdb.Expire(ctx, dayKey, 30*24*time.Hour)
		}

		// 更新 top_keys 和 top_amount ZSET
		member := strconv.FormatUint(event.KeyID, 10)
		if err := w.rdb.ZIncrBy(ctx, topKeysZSetKey(event.TenantID), 1, member).Err(); err != nil {
			zap.L().Debug("top_keys ZINCRBY failed", zap.Error(err))
		}
		if err := w.rdb.ZIncrBy(ctx, topAmountZSetKey(event.TenantID), float64(event.Amount), member).Err(); err != nil {
			zap.L().Debug("top_amount ZINCRBY failed", zap.Error(err))
		}
	}

	zap.L().Debug("ConsumeEvent 处理成功",
		zap.Uint64("key_id", event.KeyID),
		zap.Int64("remaining", event.RemainingAfter),
	)
	return nil
}

// handleAdjust 处理额度调整事件
func (w *MQWorker) handleAdjust(d amqp.Delivery) error {
	var event AdjustEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		zap.L().Error("反序列化 AdjustEvent 失败", zap.Error(err))
		return nil // 格式错误不重试
	}

	tx := w.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx: %w", tx.Error)
	}

	// 更新 key 余额（增量调整）和状态
	result := tx.Model(&model.Key{}).Where("id = ?", event.KeyID).
		Updates(map[string]interface{}{
			"remaining_amount": gorm.Expr("remaining_amount + ?", event.Delta),
			"status":           event.StatusAfter,
		})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("update key: %w", result.Error)
	}

	// 写入额度调整日志
	log := model.BalanceLog{
		TenantID:     event.TenantID,
		KeyID:        event.KeyID,
		KeyAlias:     event.KeyAlias,
		Delta:        event.Delta,
		BeforeAmount: event.RemainingAfter - event.Delta, // 反推操作前余额
		AfterAmount:  event.RemainingAfter,
		Operator:     event.Operator,
		Remark:       event.Remark,
		CreatedAt:    time.UnixMilli(event.Timestamp),
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("insert balance_log: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	zap.L().Debug("AdjustEvent 处理成功",
		zap.Uint64("key_id", event.KeyID),
		zap.Int64("delta", event.Delta),
		zap.Int64("remaining", event.RemainingAfter),
	)
	return nil
}
