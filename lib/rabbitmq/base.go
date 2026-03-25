package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageHandler 处理收到消息的函数类型
type MessageHandler func(data []byte)

// Publisher 发布者
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

// NewPublisher 创建发布者
func NewPublisher(addr string) (*Publisher, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// Publish 发布消息到指定队列
func (p *Publisher) Publish(ctx context.Context, queue string, data interface{}) error {
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

	_, err = p.channel.QueueDeclare(
		queue, // 队列名
		true,  // 持久化
		false, // 自动删除
		false, // 独占
		false, // 等待参数
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         bytes,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// PublishToExchange 发布消息到指定交换机
func (p *Publisher) PublishToExchange(ctx context.Context, exchange, routingKey string, data interface{}) error {
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

	err = p.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         bytes,
		},
	)
	return err
}

// Close 关闭发布者
func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// Subscriber 订阅者
type Subscriber struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	handlers map[string]MessageHandler
	mu      sync.Mutex
	done    chan struct{}
}

// NewSubscriber 创建订阅者
func NewSubscriber(addr string) (*Subscriber, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Subscriber{
		conn:     conn,
		channel:  ch,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]MessageHandler),
		done:     make(chan struct{}),
	}, nil
}

// Subscribe 订阅指定队列
func (s *Subscriber) Subscribe(queue string, handler MessageHandler) error {
	s.mu.Lock()
	s.handlers[queue] = handler
	s.mu.Unlock()

	_, err := s.channel.QueueDeclare(
		queue, // 队列名
		true,  // 持久化
		false, // 自动删除
		false, // 独占
		false, // 等待参数
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := s.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				s.mu.Lock()
				h, exists := s.handlers[queue]
				s.mu.Unlock()
				if exists && h != nil {
					h(msg.Body)
				}
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// SubscribeToExchange 订阅交换机
func (s *Subscriber) SubscribeToExchange(exchange, queue, routingKey string, handler MessageHandler) error {
	s.mu.Lock()
	s.handlers[queue] = handler
	s.mu.Unlock()

	// 声明交换机
	err := s.channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// 声明队列
	_, err = s.channel.QueueDeclare(
		queue,  // queue name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,   // no-wait
		nil,    // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 绑定队列到交换机
	err = s.channel.QueueBind(
		queue,     // queue name
		routingKey, // routing key
		exchange,  // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := s.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				s.mu.Lock()
				h, exists := s.handlers[queue]
				s.mu.Unlock()
				if exists && h != nil {
					h(msg.Body)
				}
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// SubscribeMany 订阅多个队列
func (s *Subscriber) SubscribeMany(queues []string, handler MessageHandler) error {
	for _, queue := range queues {
		if err := s.Subscribe(queue, handler); err != nil {
			return err
		}
	}
	return nil
}

// Unsubscribe 取消订阅
func (s *Subscriber) Unsubscribe(queue string) {
	s.mu.Lock()
	delete(s.handlers, queue)
	s.mu.Unlock()
}

// Close 关闭订阅者
func (s *Subscriber) Close() error {
	s.cancel()
	s.wg.Wait()
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
