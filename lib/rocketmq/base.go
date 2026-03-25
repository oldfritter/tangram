package rocketmq

import (
	"context"
	"encoding/json"
	"fmt"
)

// MessageHandler 处理收到消息的函数类型
type MessageHandler func(data []byte)

// Publisher 发布者
type Publisher struct {
	nameServer string
}

// NewPublisher 创建发布者
// nameServer: RocketMQ NameServer 地址，如 localhost:9876
func NewPublisher(nameServer string) (*Publisher, error) {
	return &Publisher{nameServer: nameServer}, nil
}

// Publish 发布消息到指定主题
func (p *Publisher) Publish(ctx context.Context, topic string, data interface{}) error {
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

	// TODO: 实现实际的 RocketMQ 发布逻辑
	// 使用 rocketmq-client-go 或其他客户端库
	fmt.Printf("[RocketMQ] Publishing to topic %s: %s\n", topic, string(bytes))
	return nil
}

// PublishWithTag 发布消息到指定主题和标签
func (p *Publisher) PublishWithTag(ctx context.Context, topic, tag string, data interface{}) error {
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

	// TODO: 实现带标签的发布逻辑
	fmt.Printf("[RocketMQ] Publishing to topic %s, tag %s: %s\n", topic, tag, string(bytes))
	return nil
}

// Close 关闭发布者
func (p *Publisher) Close() error {
	// TODO: 清理资源
	return nil
}

// Subscriber 订阅者
type Subscriber struct {
	nameServer string
	groupID    string
	topics     map[string]MessageHandler
}

// NewSubscriber 创建订阅者
// nameServer: RocketMQ NameServer 地址，如 localhost:9876
// groupID: 消费者组 ID
func NewSubscriber(nameServer string, groupID string) (*Subscriber, error) {
	return &Subscriber{
		nameServer: nameServer,
		groupID:    groupID,
		topics:     make(map[string]MessageHandler),
	}, nil
}

// Subscribe 订阅指定主题
func (s *Subscriber) Subscribe(topic string, handler MessageHandler) error {
	s.topics[topic] = handler

	// TODO: 实现实际的 RocketMQ 订阅逻辑
	// 启动消费者并注册消息处理函数
	fmt.Printf("[RocketMQ] Subscribed to topic: %s\n", topic)
	return nil
}

// SubscribeWithTag 订阅指定主题和标签
func (s *Subscriber) SubscribeWithTag(topic, tag string, handler MessageHandler) error {
	// TODO: 实现带标签的订阅逻辑
	fmt.Printf("[RocketMQ] Subscribed to topic: %s, tag: %s\n", topic, tag)
	s.topics[topic] = handler
	return nil
}

// SubscribeMany 订阅多个主题
func (s *Subscriber) SubscribeMany(topics []string, handler MessageHandler) error {
	for _, topic := range topics {
		if err := s.Subscribe(topic, handler); err != nil {
			return err
		}
	}
	return nil
}

// Unsubscribe 取消订阅
func (s *Subscriber) Unsubscribe(topic string) {
	delete(s.topics, topic)
	// TODO: 取消订阅逻辑
}

// Close 关闭订阅者
func (s *Subscriber) Close() error {
	// TODO: 停止消费者并清理资源
	return nil
}

// Ping 检查连接
func (p *Publisher) Ping(ctx context.Context) error {
	// TODO: 实现连接检查
	return nil
}

// Ping 检查连接
func (s *Subscriber) Ping(ctx context.Context) error {
	// TODO: 实现连接检查
	return nil
}

// HelperPublish 辅助方法：发布消息（别名）
func HelperPublish(nameServer string, topic string, data interface{}) error {
	p, err := NewPublisher(nameServer)
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Publish(context.Background(), topic, data)
}
