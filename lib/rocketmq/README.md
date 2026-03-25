# RocketMQ 使用指南

本模块提供基于 RocketMQ 的消息通讯功能。

## 快速开始

### 1. 安装依赖

```bash
go get github.com/apache/rocketmq-client-go/v2
```

### 2. 基本用法

#### 发布消息

```go
package main

import (
	"context"
	"fmt"

	"github.com/oldfritter/tangram/lib/rocketmq"
)

func main() {
	ctx := context.Background()

	// 创建生产者
	pub, err := rocketmq.NewPublisher("localhost:9876")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer pub.Close()

	// 发布消息
	err = pub.Publish(ctx, "my_topic", map[string]interface{}{
		"type": "greeting",
		"msg":  "Hello!",
	})
	if err != nil {
		fmt.Println("Publish error:", err)
	}
}
```

#### 订阅消息

```go
package main

import (
	"fmt"
	"time"

	"github.com/oldfritter/tangram/lib/rocketmq"
)

func main() {
	// 创建消费者
	sub, err := rocketmq.NewSubscriber("localhost:9876", "my_group")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer sub.Close()

	// 订阅主题
	err = sub.Subscribe("my_topic", func(data []byte) {
		fmt.Println("Received:", string(data))
	})
	if err != nil {
		fmt.Println("Subscribe error:", err)
	}

	// 保持程序运行
	time.Sleep(time.Second * 10)
}
```

### 3. 完整示例

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/oldfritter/tangram/lib/rocketmq"
)

// 自定义消息结构
type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	ctx := context.Background()

	// 启动消费者
	consumer, _ := rocketmq.NewSubscriber("localhost:9876", "demo_group")
	defer consumer.Close()

	consumer.Subscribe("demo_topic", func(data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("Received: [Type=%s] %s\n", msg.Type, msg.Content)
	})

	time.Sleep(500 * time.Millisecond)

	// 启动生产者
	producer, _ := rocketmq.NewPublisher("localhost:9876")
	defer producer.Close()

	for i := 0; i < 5; i++ {
		msg := Message{
			Type:      "test",
			Content:   fmt.Sprintf("Message %d", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := producer.Publish(ctx, "demo_topic", msg); err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("Published message %d\n", i+1)
		time.Sleep(time.Second)
	}
}
```

## API 参考

### Publisher

| 方法 | 说明 |
|------|------|
| `NewPublisher(nameServer, options)` | 创建生产者实例 |
| `Publish(ctx, topic, data)` | 发布消息到指定主题 |
| `PublishWithTag(ctx, topic, tag, data)` | 发布消息到指定主题和标签 |
| `Close()` | 关闭连接 |

### Subscriber

| 方法 | 说明 |
|------|------|
| `NewSubscriber(nameServer, groupID, options)` | 创建订阅者实例 |
| `Subscribe(topic, handler)` | 订阅主题 |
| `SubscribeWithTag(topic, tag, handler)` | 订阅带标签的主题 |
| `SubscribeMany(topics, handler)` | 订阅多个主题 |
| `Close()` | 关闭连接 |

## 配置说明

### NameServer 地址

RocketMQ NameServer 是轻量级的注册中心，生产者和消费者通过它发现 Broker。

格式：`host:port`，如 `localhost:9876`

### 消费者组

消费者组是一组消费者的标识，同一组内的消费者共同消费主题消息。

## 与其他 MQ 对比

| 特性 | Redis Pub/Sub | Kafka | RabbitMQ | RocketMQ |
|------|---------------|-------|----------|----------|
| 延迟 | 低 | 中 | 低 | 低 |
| 持久化 | ❌ | ✅ | ✅ | ✅ |
| 消息确认 | ❌ | ✅ | ✅ | ✅ |
| 消费者组 | ❌ | ✅ | ✅ | ✅ |
| 事务消息 | ❌ | ❌ | ✅ | ✅ |
| 顺序消息 | ✅ | ✅ | ✅ | ✅ |
| 延迟消息 | ❌ | ❌ | ✅ | ✅ |

## 注意事项

1. **NameServer**：RocketMQ 使用 NameServer 作为服务发现中心
2. **消费者组**：同一消费者组内的消费者共享消费进度
3. **线程安全**：Producer 支持并发发布
4. **消息顺序**：RocketMQ 支持顺序消息投递
