package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// MessageHandler 处理收到消息的函数类型
type MessageHandler func(data []byte)

// Producer 生产者
type Producer struct {
	producer sarama.SyncProducer
	mu       sync.Mutex
}

// NewProducer 创建生产者
func NewProducer(addrs []string, config *sarama.Config) (*Producer, error) {
	if config == nil {
		config = sarama.NewConfig()
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Retry.Max = 3
		config.Producer.Return.Successes = true
	}

	producer, err := sarama.NewSyncProducer(addrs, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{producer: producer}, nil
}

// Publish 发布消息到指定主题
func (p *Producer) Publish(ctx context.Context, topic string, key string, data interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var bytes []byte
	var err error

	switch v := data.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		bytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	if key != "" {
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// PublishAsync 异步发布消息
func (p *Producer) PublishAsync(topic string, key string, data interface{}) error {
	var bytes []byte
	var err error

	switch v := data.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		bytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.producer.Close()
}

// Consumer 消费者
type Consumer struct {
	consumer sarama.ConsumerGroup
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	handlers map[string]MessageHandler
	mu       sync.Mutex
}

// NewConsumer 创建消费者
func NewConsumer(addrs []string, groupID string, config *sarama.Config) (*Consumer, error) {
	if config == nil {
		config = sarama.NewConfig()
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	consumer, err := sarama.NewConsumerGroup(addrs, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		consumer: consumer,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]MessageHandler),
	}, nil
}

// ConsumeHandler 实现 sarama.ConsumerGroupHandler
type ConsumeHandler struct {
	handlers map[string]MessageHandler
	mu       *sync.Mutex
}

// Setup 是消费者会话开始的回调
func (h *ConsumeHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 是消费者会话结束的回调
func (h *ConsumeHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消费到的消息
func (h *ConsumeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.mu.Lock()
		handler, exists := h.handlers[msg.Topic]
		h.mu.Unlock()

		if exists && handler != nil {
			handler(msg.Value)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

// Subscribe 订阅指定主题
func (c *Consumer) Subscribe(topic string, handler MessageHandler) {
	c.mu.Lock()
	c.handlers[topic] = handler
	c.mu.Unlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		handler := &ConsumeHandler{
			handlers: c.handlers,
			mu:       &c.mu,
		}

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				if err := c.consumer.Consume(c.ctx, []string{topic}, handler); err != nil {
					time.Sleep(time.Second)
				}
			}
		}
	}()
}

// SubscribeMany 订阅多个主题
func (c *Consumer) SubscribeMany(topics []string, handler MessageHandler) {
	for _, topic := range topics {
		c.Subscribe(topic, handler)
	}
}

// Unsubscribe 取消订阅
func (c *Consumer) Unsubscribe(topic string) {
	c.mu.Lock()
	delete(c.handlers, topic)
	c.mu.Unlock()
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.cancel()
	c.wg.Wait()
	return c.consumer.Close()
}
