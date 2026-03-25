package tangram

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/IBM/sarama"
	"github.com/oldfritter/tangram/lib/kafka"
	"github.com/oldfritter/tangram/lib/rabbitmq"
	"github.com/oldfritter/tangram/lib/redis"
	"github.com/oldfritter/tangram/lib/rocketmq"
	"github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v3"
)

// MessageHandler 处理收到消息的函数类型
type MessageHandler func(data []byte)

// MQConfig MQ 配置结构（对应 example/config/app.yml）
type MQConfig struct {
	// 类型: kafka, rabbitmq, redis, rocketmq
	Type string `yaml:"type"`

	// Redis 配置
	Redis RedisConfig `yaml:"redis"`

	// Kafka 配置
	Kafka KafkaConfig `yaml:"kafka"`

	// RabbitMQ 配置
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`

	// RocketMQ 配置
	RocketMQ RocketMQConfig `yaml:"rocketmq"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`     // 地址，如 localhost:6379
	Password string `yaml:"password"` // 密码
	DB       int    `yaml:"db"`       // 数据库编号
}

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Addrs   []string `yaml:"addrs"`   // Broker 地址列表
	GroupID string   `yaml:"groupId"` // 消费者组 ID
}

// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
	Addr string `yaml:"addr"` // 连接地址，如 amqp://guest:guest@localhost:5672/
}

// RocketMQConfig RocketMQ 配置
type RocketMQConfig struct {
	NameServer string `yaml:"nameServer"` // NameServer 地址，如 localhost:9876
	GroupID    string `yaml:"groupId"`    // 消费者组 ID
}

// publisher 发布者接口
type publisher interface {
	Publish(ctx context.Context, topic string, data interface{}) error
	Close() error
}

// subscriber 订阅者接口
type subscriber interface {
	Subscribe(topic string, handler MessageHandler) error
	SubscribeMany(topics []string, handler MessageHandler) error
	Unsubscribe(topic string)
	Close() error
}

// MQ 消息队列统一入口
type MQ struct {
	pub     publisher
	sub     subscriber
	cfg     *MQConfig
	mu      sync.Mutex
	msgType string // 当前消息类型：kafka=topic, rabbitmq=queue, redis=channel
}

// NewMQ 根据配置创建 MQ 实例
func NewMQ(cfg *MQConfig) (*MQ, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	mq := &MQ{cfg: cfg}

	switch cfg.Type {
	case "redis":
		mq.msgType = "channel"
		if err := mq.initRedis(); err != nil {
			return nil, fmt.Errorf("init redis failed: %w", err)
		}
	case "kafka":
		mq.msgType = "topic"
		if err := mq.initKafka(); err != nil {
			return nil, fmt.Errorf("init kafka failed: %w", err)
		}
	case "rabbitmq":
		mq.msgType = "queue"
		if err := mq.initRabbitMQ(); err != nil {
			return nil, fmt.Errorf("init rabbitmq failed: %w", err)
		}
	case "rocketmq":
		mq.msgType = "topic"
		if err := mq.initRocketMQ(); err != nil {
			return nil, fmt.Errorf("init rocketmq failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported MQ type: %s", cfg.Type)
	}

	return mq, nil
}

// LoadConfigFromYAML 从 YAML 配置文件加载配置
// 自动读取 config/app.yml 文件
func LoadConfigFromYAML(path string) (*MQConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}

	var cfg MQConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	// 验证配置
	if cfg.Type == "" {
		return nil, fmt.Errorf("mq.type is required in config")
	}

	return &cfg, nil
}

// LoadDefaultConfig 加载默认位置的配置文件 (example/config/app.yml)
func LoadDefaultConfig() (*MQConfig, error) {
	return LoadConfigFromYAML("example/config/app.yml")
}

// LoadConfigFromYAMLString 从 YAML 字符串加载配置
func LoadConfigFromYAMLString(yamlStr string) (*MQConfig, error) {
	var cfg MQConfig
	if err := yaml.Unmarshal([]byte(yamlStr), &cfg); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	if cfg.Type == "" {
		return nil, fmt.Errorf("mq.type is required in config")
	}

	return &cfg, nil
}

func (m *MQ) initRedis() error {
	m.pub = redis.NewPublisher(
		m.cfg.Redis.Addr,
		m.cfg.Redis.Password,
		m.cfg.Redis.DB,
	)
	m.sub = redis.NewSubscriber(
		m.cfg.Redis.Addr,
		m.cfg.Redis.Password,
		m.cfg.Redis.DB,
	)
	return nil
}

func (m *MQ) initKafka() error {
	var err error

	// 生产者
	m.pub, err = kafka.NewProducer(m.cfg.Kafka.Addrs, nil)
	if err != nil {
		return err
	}

	// 消费者
	m.sub, err = kafka.NewConsumer(m.cfg.Kafka.Addrs, m.cfg.Kafka.GroupID, nil)
	if err != nil {
		m.pub.(*kafka.Producer).Close()
		return err
	}

	return nil
}

func (m *MQ) initRabbitMQ() error {
	var err error

	// 发布者
	m.pub, err = rabbitmq.NewPublisher(m.cfg.RabbitMQ.Addr)
	if err != nil {
		return err
	}

	// 订阅者
	m.sub, err = rabbitmq.NewSubscriber(m.cfg.RabbitMQ.Addr)
	if err != nil {
		m.pub.(*rabbitmq.Publisher).Close()
		return err
	}

	return nil
}

func (m *MQ) initRocketMQ() error {
	var err error

	// 生产者
	m.pub, err = rocketmq.NewPublisher(m.cfg.RocketMQ.NameServer)
	if err != nil {
		return err
	}

	// 消费者
	m.sub, err = rocketmq.NewSubscriber(m.cfg.RocketMQ.NameServer, m.cfg.RocketMQ.GroupID)
	if err != nil {
		m.pub.(*rocketmq.Publisher).Close()
		return err
	}

	return nil
}

// Publish 发布消息
// topic 的含义取决于配置的 MQ 类型:
// - Redis: channel 名称
// - Kafka: topic 名称
// - RabbitMQ: queue 名称
// - RocketMQ: topic 名称
func (m *MQ) Publish(ctx context.Context, topic string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pub == nil {
		return fmt.Errorf("publisher not initialized")
	}

	switch m.cfg.Type {
	case "kafka":
		// Kafka 需要 key 参数，这里使用空字符串
		return m.pub.Publish(ctx, topic, "", data)
	default:
		return m.pub.Publish(ctx, topic, data)
	}
}

// Subscribe 订阅消息
// topic 的含义取决于配置的 MQ 类型:
// - Redis: channel 名称
// - Kafka: topic 名称
// - RabbitMQ: queue 名称
// - RocketMQ: topic 名称
func (m *MQ) Subscribe(topic string, handler MessageHandler) error {
	if m.sub == nil {
		return fmt.Errorf("subscriber not initialized")
	}

	switch m.cfg.Type {
	case "kafka":
		m.sub.Subscribe(topic, handler)
		return nil
	default:
		return m.sub.Subscribe(topic, handler)
	}
}

// SubscribeMany 订阅多个主题
func (m *MQ) SubscribeMany(topics []string, handler MessageHandler) error {
	if m.sub == nil {
		return fmt.Errorf("subscriber not initialized")
	}

	switch m.cfg.Type {
	case "kafka":
		m.sub.SubscribeMany(topics, handler)
		return nil
	default:
		return m.sub.SubscribeMany(topics, handler)
	}
}

// Unsubscribe 取消订阅
func (m *MQ) Unsubscribe(topic string) {
	if m.sub == nil {
		return
	}
	m.sub.Unsubscribe(topic)
}

// Close 关闭连接
func (m *MQ) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	if m.pub != nil {
		if err := m.pub.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if m.sub != nil {
		if err := m.sub.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	m.pub = nil
	m.sub = nil

	if len(errs) > 0 {
		return fmt.Errorf("close failed: %v", errs)
	}
	return nil
}

// GetType 获取当前 MQ 类型
func (m *MQ) GetType() string {
	return m.cfg.Type
}

// GetMsgType 获取消息类型名称
func (m *MQ) GetMsgType() string {
	return m.msgType
}

// ==================== 类型断言需要的接口 ====================

// RedisPublisherExtension Redis 发布者扩展接口
type RedisPublisherExtension interface {
	Publish(ctx context.Context, channel string, data interface{}) error
	Close() error
}

// RedisSubscriberExtension Redis 订阅者扩展接口
type RedisSubscriberExtension interface {
	Subscribe(channel string, handler MessageHandler)
	SubscribeMany(channels []string, handler MessageHandler)
	Unsubscribe(channel string)
	Close() error
}

// KafkaPublisherExtension Kafka 生产者扩展接口
type KafkaPublisherExtension interface {
	Publish(ctx context.Context, topic string, key string, data interface{}) error
	PublishAsync(topic string, key string, data interface{}) error
	Close() error
}

// KafkaSubscriberExtension Kafka 消费者扩展接口
type KafkaSubscriberExtension interface {
	Subscribe(topic string, handler MessageHandler)
	SubscribeMany(topics []string, handler MessageHandler)
	Unsubscribe(topic string)
	Close() error
}

// RabbitMQPublisherExtension RabbitMQ 发布者扩展接口
type RabbitMQPublisherExtension interface {
	Publish(ctx context.Context, queue string, data interface{}) error
	PublishToExchange(ctx context.Context, exchange, routingKey string, data interface{}) error
	Close() error
}

// RabbitMQSubscriberExtension RabbitMQ 订阅者扩展接口
type RabbitMQSubscriberExtension interface {
	Subscribe(queue string, handler MessageHandler) error
	SubscribeToExchange(exchange, queue, routingKey string, handler MessageHandler) error
	SubscribeMany(queues []string, handler MessageHandler) error
	Unsubscribe(queue string)
	Close() error
}

// ==================== 兼容性类型别名 ====================

// KafkaConfigAlias Kafka 配置别名（兼容 sarama）
type KafkaConfigAlias = sarama.Config

// RabbitMQAddress RabbitMQ 地址格式别名
type RabbitMQAddress = amqp091.Connection

// ExampleYAMLConfig 示例配置 YAML
const ExampleYAMLConfig = `mq:
  type: "kafka"
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
  kafka:
    addrs:
      - "localhost:9092"
    groupId: "my_group"
  rabbitmq:
    addr: "amqp://guest:guest@localhost:5672/"
`
