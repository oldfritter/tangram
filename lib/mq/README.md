# 统一消息队列 (MQ) 使用指南

本模块提供统一的消息队列接口，根据配置自动选择底层实现（Redis、Kafka 或 RabbitMQ）。

## 快速开始

### 1. 配置方式

#### 方式一：从文件加载配置

创建 `mq_config.json`：

```json
{
  "type": "kafka",
  "redis": {
    "addr": "localhost:6379",
    "password": "",
    "db": 0
  },
  "kafka": {
    "addrs": ["localhost:9092"],
    "groupId": "my_group"
  },
  "rabbitmq": {
    "addr": "amqp://guest:guest@localhost:5672/"
  }
}
```

代码中加载：

```go
cfg, err := mq.LoadConfig("mq_config.json")
if err != nil {
    log.Fatal(err)
}

mqInstance, err := mq.NewMQ(cfg)
if err != nil {
    log.Fatal(err)
}
defer mqInstance.Close()
```

#### 方式二：从 JSON 字符串加载

```go
cfg, err := mq.LoadConfigFromJSON(`{"type": "kafka", "kafka": {"addrs": ["localhost:9092"], "groupId": "my_group"}}`)
```

### 2. 使用示例

#### 发布消息

```go
ctx := context.Background()

err := mqInstance.Publish(ctx, "my_topic", map[string]interface{}{
    "type": "greeting",
    "msg":  "Hello!",
})
if err != nil {
    fmt.Println("Publish error:", err)
}
```

#### 订阅消息

```go
err := mqInstance.Subscribe("my_topic", func(data []byte) {
    fmt.Println("Received:", string(data))
})
if err != nil {
    fmt.Println("Subscribe error:", err)
}
```

### 3. 配置说明

| 类型 | topic 参数含义 | 配置字段 |
|------|----------------|----------|
| Redis | channel 名称 | `redis.addr`, `redis.password`, `redis.db` |
| Kafka | topic 名称 | `kafka.addrs`, `kafka.groupId` |
| RabbitMQ | queue 名称 | `rabbitmq.addr` |

### 4. 完整示例

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/oldfritter/tangram/lib/mq"
)

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	// 根据环境变量选择 MQ 类型
	mqType := "kafka" // 可选: redis, kafka, rabbitmq

	var cfg *mq.Config
	switch mqType {
	case "redis":
		cfg, _ = mq.LoadConfigFromJSON(`{
			"type": "redis",
			"redis": {"addr": "localhost:6379", "password": "", "db": 0}
		}`)
	case "kafka":
		cfg, _ = mq.LoadConfigFromJSON(`{
			"type": "kafka",
			"kafka": {"addrs": ["localhost:9092"], "groupId": "demo_group"}
		}`)
	case "rabbitmq":
		cfg, _ = mq.LoadConfigFromJSON(`{
			"type": "rabbitmq",
			"rabbitmq": {"addr": "amqp://guest:guest@localhost:5672/"}
		}`)
	}

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

## API 参考

| 方法 | 说明 |
|------|------|
| `NewMQ(cfg)` | 根据配置创建 MQ 实例 |
| `LoadConfig(path)` | 从文件加载配置 |
| `LoadConfigFromJSON(json)` | 从 JSON 字符串加载配置 |
| `Publish(ctx, topic, data)` | 发布消息 |
| `Subscribe(topic, handler)` | 订阅消息 |
| `SubscribeMany(topics, handler)` | 订阅多个主题 |
| `Unsubscribe(topic)` | 取消订阅 |
| `Close()` | 关闭连接 |
| `GetType()` | 获取当前 MQ 类型 |
| `GetMsgType()` | 获取消息类型名称 |

## 切换 MQ 方案

只需修改配置中的 `type` 字段，无需修改业务代码：

| type 值 | 底层实现 |
|---------|----------|
| `"redis"` | Redis Pub/Sub |
| `"kafka"` | Kafka Producer/Consumer |
| `"rabbitmq"` | RabbitMQ 队列 |

## 进阶使用

### 访问底层实现

如需使用特定 MQ 的高级特性，可以通过类型断言：

```go
// 访问 Kafka 生产者
if mqInstance.GetType() == "kafka" {
    // 获取底层 producer 进行高级操作
    // 注意：需要自行维护类型
}
```

### RabbitMQ 交换机模式

RabbitMQ 支持交换机模式，但统一接口仅封装了队列模式。如需使用交换机，请直接使用 `rabbitmq` 包：

```go
sub, _ := rabbitmq.NewSubscriber("amqp://guest:guest@localhost:5672/")
sub.SubscribeToExchange("my_exchange", "my_queue", "order.*", handler)
```
