# 统一消息队列 (MQ) 使用指南

> [English](./MQ.md)

本模块提供统一的消息队列接口，根据配置文件自动选择底层实现（Redis、Kafka、RabbitMQ 或 RocketMQ）。

## 快速开始

### 1. 配置 MQ

在 `example/config/app.yml` 中配置使用的 MQ 类型：

```yaml
mq:
  type: "kafka"  # 可选: redis, kafka, rabbitmq, rocketmq

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

  rocketmq:
    nameServer: "localhost:9876"
    groupId: "my_group"
```

### 2. 使用示例

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/oldfritter/tangram"
)

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	// 自动读取 example/config/app.yml
	cfg, err := mq.LoadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 根据配置自动选择 MQ 实现
	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	ctx := context.Background()
	topic := "demo_topic"

	// 启动消费者
	mqInstance.Subscribe(topic, func(data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("Received: [Type=%s] %s\n", msg.Type, msg.Content)
	})

	time.Sleep(500 * time.Millisecond)

	// 启动生产者
	for i := 0; i < 5; i++ {
		msg := Message{
			Type:      "test",
			Content:   fmt.Sprintf("Message %d", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := mqInstance.Publish(ctx, topic, msg); err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("Published message %d\n", i+1)
		time.Sleep(time.Second)
	}
}
```

## 配置说明

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `mq.type` | MQ 类型 | `kafka`, `redis`, `rabbitmq` |
| `mq.redis.addr` | Redis 地址 | `localhost:6379` |
| `mq.redis.password` | Redis 密码 | `""` |
| `mq.redis.db` | Redis 数据库编号 | `0` |
| `mq.kafka.addrs` | Kafka Broker 地址 | `["localhost:9092"]` |
| `mq.kafka.groupId` | 消费者组 ID | `my_group` |
| `mq.rabbitmq.addr` | RabbitMQ 地址 | `amqp://guest:guest@localhost:5672/` |
| `mq.rocketmq.nameServer` | RocketMQ NameServer 地址 | `localhost:9876` |
| `mq.rocketmq.groupId` | RocketMQ 消费者组 ID | `my_group` |

## API 参考

| 方法 | 说明 |
|------|------|
| `LoadDefaultConfig()` | 自动加载 `example/config/app.yml` |
| `LoadConfigFromYAML(path)` | 加载指定路径的 YAML 配置 |
| `LoadConfigFromYAMLString(yaml)` | 从 YAML 字符串加载配置 |
| `NewMQ(cfg)` | 根据配置创建 MQ 实例 |
| `Publish(ctx, topic, data)` | 发布消息 |
| `Subscribe(topic, handler)` | 订阅消息 |
| `SubscribeMany(topics, handler)` | 订阅多个主题 |
| `Unsubscribe(topic)` | 取消订阅 |
| `Close()` | 关闭连接 |
| `GetType()` | 获取当前 MQ 类型 |
| `GetMsgType()` | 获取消息类型名称 |

## 切换 MQ 方案

只需修改 `example/config/app.yml` 中的 `mq.type` 字段，业务代码无需任何改动：

| type 值 | 底层实现 |
|---------|----------|
| `"redis"` | Redis Pub/Sub |
| `"kafka"` | Kafka Producer/Consumer |
| `"rabbitmq"` | RabbitMQ 队列 |
| `"rocketmq"` | RocketMQ Producer/Consumer |

## topic 参数含义

| MQ 类型 | topic 参数含义 |
|---------|----------------|
| Redis | channel 名称 |
| Kafka | topic 名称 |
| RabbitMQ | queue 名称 |
| RocketMQ | topic 名称 |

## 项目结构

```
your-project/
├── example/
│   ├── config/
│   │   └── app.yml     # MQ 配置文件
│   ├── redis/
│   ├── kafka/
│   ├── rabbitmq/
│   └── rocketmq/
├── mq.go               # 统一 MQ 接口
└── main.go             # 业务代码
```
