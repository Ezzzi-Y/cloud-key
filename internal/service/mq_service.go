package service

import (
	"CloudKey/internal/config"
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	mqExchange     = "cloudkey.events"
	mqConsumeQueue = "cloudkey.consume"
	mqAdjustQueue  = "cloudkey.adjust"
	mqDLQQueue     = "cloudkey.dlq"
	mqPrefetch     = 10
	mqMaxRetries   = 3
)

// ConsumeEvent 消费扣减事件
type ConsumeEvent struct {
	EventID        string `json:"event_id"`
	RequestID      string `json:"request_id"`
	KeyID          uint64 `json:"key_id"`
	KeyAlias       string `json:"key_alias"`
	KeySuffix      string `json:"key_suffix"`
	TenantID       uint64 `json:"tenant_id"`
	Amount         int64  `json:"amount"`
	RemainingAfter int64  `json:"remaining_after"`
	StatusAfter    string `json:"status_after"`
	IP             string `json:"ip"`
	UserAgent      string `json:"user_agent"`
	Timestamp      int64  `json:"timestamp"`
	UsedAt         int64  `json:"used_at"`
}

// AdjustEvent 额度调整事件
type AdjustEvent struct {
	EventID        string `json:"event_id"`
	RequestID      string `json:"request_id"`
	KeyID          uint64 `json:"key_id"`
	KeyAlias       string `json:"key_alias"`
	KeySuffix      string `json:"key_suffix"`
	TenantID       uint64 `json:"tenant_id"`
	Delta          int64  `json:"delta"`
	RemainingAfter int64  `json:"remaining_after"`
	StatusAfter    string `json:"status_after"`
	Operator       string `json:"operator"`
	Remark         string `json:"remark"`
	Timestamp      int64  `json:"timestamp"`
}

// MQService RabbitMQ 连接管理与消息发布
type MQService struct {
	conn           *amqp.Connection
	consumeChannel *amqp.Channel
	adjustChannel  *amqp.Channel
}

// NewMQService 建立 RabbitMQ 连接，声明 exchange、业务队列、DLQ 队列
func NewMQService(cfg config.MQConfig) (*MQService, error) {
	vhost := cfg.VHost
	if vhost == "" {
		vhost = "/"
	}
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, vhost)
	zap.L().Info("连接 RabbitMQ",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("username", cfg.Username),
		zap.String("vhost", vhost),
		zap.String("url", url),
	)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to RabbitMQ: %w", err)
	}

	consumeCh, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open consume channel: %w", err)
	}

	adjustCh, err := conn.Channel()
	if err != nil {
		consumeCh.Close()
		conn.Close()
		return nil, fmt.Errorf("open adjust channel: %w", err)
	}

	svc := &MQService{
		conn:           conn,
		consumeChannel: consumeCh,
		adjustChannel:  adjustCh,
	}

	if err := svc.setupTopology(); err != nil {
		svc.Close()
		return nil, fmt.Errorf("setup topology: %w", err)
	}

	zap.L().Info("RabbitMQ 连接成功", zap.String("url", url))
	return svc, nil
}

// setupTopology 声明 exchange、队列和绑定关系
func (s *MQService) setupTopology() error {
	// 声明 exchange
	if err := s.consumeChannel.ExchangeDeclare(
		mqExchange, "topic", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	// 声明 DLQ 队列
	dlqArgs := amqp.Table{
		"x-queue-type": "quorum",
	}
	if _, err := s.consumeChannel.QueueDeclare(
		mqDLQQueue, true, false, false, false, dlqArgs,
	); err != nil {
		return fmt.Errorf("declare DLQ queue: %w", err)
	}

	// 声明 consume 队列（带 DLQ 死信路由）
	consumeArgs := amqp.Table{
		"x-dead-letter-exchange":    mqExchange,
		"x-dead-letter-routing-key": "dlq",
		"x-queue-type":              "quorum",
	}
	if _, err := s.consumeChannel.QueueDeclare(
		mqConsumeQueue, true, false, false, false, consumeArgs,
	); err != nil {
		return fmt.Errorf("declare consume queue: %w", err)
	}

	// 声明 adjust 队列（带 DLQ 死信路由）
	adjustArgs := amqp.Table{
		"x-dead-letter-exchange":    mqExchange,
		"x-dead-letter-routing-key": "dlq",
		"x-queue-type":              "quorum",
	}
	if _, err := s.adjustChannel.QueueDeclare(
		mqAdjustQueue, true, false, false, false, adjustArgs,
	); err != nil {
		return fmt.Errorf("declare adjust queue: %w", err)
	}

	// 绑定队列到 exchange
	bindings := []struct {
		queue   string
		routing string
	}{
		{mqConsumeQueue, "consume.*"},
		{mqAdjustQueue, "adjust.*"},
		{mqDLQQueue, "dlq"},
	}

	ch := s.consumeChannel // 用同一个 channel 做绑定
	for _, b := range bindings {
		if err := ch.QueueBind(b.queue, b.routing, mqExchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %s: %w", b.queue, err)
		}
	}

	return nil
}

// PublishConsumeEvent 发布消费扣减事件
func (s *MQService) PublishConsumeEvent(event ConsumeEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal consume event: %w", err)
	}

	routingKey := fmt.Sprintf("consume.%d", event.TenantID)
	return s.consumeChannel.PublishWithContext(
		context.Background(),
		mqExchange,
		routingKey,
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.EventID,
			Body:         body,
		},
	)
}

// PublishAdjustEvent 发布额度调整事件
func (s *MQService) PublishAdjustEvent(event AdjustEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal adjust event: %w", err)
	}

	routingKey := fmt.Sprintf("adjust.%d", event.TenantID)
	return s.adjustChannel.PublishWithContext(
		context.Background(),
		mqExchange,
		routingKey,
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.EventID,
			Body:         body,
		},
	)
}

// Close 关闭所有 channel 和连接
func (s *MQService) Close() {
	if s.consumeChannel != nil {
		s.consumeChannel.Close()
	}
	if s.adjustChannel != nil {
		s.adjustChannel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
