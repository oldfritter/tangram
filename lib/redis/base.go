package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// MessageHandler 处理收到消息的函数类型
type MessageHandler func(data []byte)

// Publisher 发布者
type Publisher struct {
	client *redis.Client
	mu     sync.Mutex
}

// NewPublisher 创建发布者
func NewPublisher(addr string, password string, db int) *Publisher {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Publisher{client: client}
}

// Publish 发布消息到指定频道
func (p *Publisher) Publish(ctx context.Context, channel string, data interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var bytes []byte
	switch v := data.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		var err error
		bytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	return p.client.Publish(ctx, channel, bytes).Err()
}

// Close 关闭发布者
func (p *Publisher) Close() error {
	return p.client.Close()
}

// Subscriber 订阅者
type Subscriber struct {
	client   *redis.Client
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	handlers map[string]MessageHandler
	mu       sync.Mutex
}

// NewSubscriber 创建订阅者
func NewSubscriber(addr string, password string, db int) *Subscriber {
	ctx, cancel := context.WithCancel(context.Background())
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Subscriber{
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]MessageHandler),
	}
}

// Subscribe 订阅指定频道
func (s *Subscriber) Subscribe(channel string, handler MessageHandler) {
	s.mu.Lock()
	s.handlers[channel] = handler
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		pubsub := s.client.Subscribe(s.ctx, channel)
		ch := pubsub.Channel()

		for {
			select {
			case <-s.ctx.Done():
				pubsub.Close()
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				s.mu.Lock()
				h, exists := s.handlers[channel]
				s.mu.Unlock()
				if exists && h != nil {
					h([]byte(msg.Payload))
				}
			}
		}
	}()
}

// SubscribeMany 订阅多个频道
func (s *Subscriber) SubscribeMany(channels []string, handler MessageHandler) {
	for _, ch := range channels {
		s.Subscribe(ch, handler)
	}
}

// Unsubscribe 取消订阅
func (s *Subscriber) Unsubscribe(channel string) {
	s.mu.Lock()
	delete(s.handlers, channel)
	s.mu.Unlock()
}

// Close 关闭订阅者
func (s *Subscriber) Close() error {
	s.cancel()
	s.wg.Wait()
	return s.client.Close()
}

// Ping 检查连接
func (p *Publisher) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}

// Ping 检查连接
func (s *Subscriber) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
