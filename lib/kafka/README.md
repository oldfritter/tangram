# Kafka 使用指南

本模块提供基于 Kafka 的消息通讯功能，支持发布/订阅模式。

## 快速开始

### 1. 安装依赖

```bash
go get github.com/IBM/sarama
```

### 2. 基本用法

#### 发布消息

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/oldfritter/tangram/lib/kafka"
)

func main() {
	ctx := context.Background()
	
	// 创建生产者
	producer, err := kafka.NewProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer producer.Close()
	
	// 发布消息
	err = producer.Publish(ctx, "my_topic", "key", map[string]interface{}{
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
	
	"github.com/oldfritter/tangram/lib/kafka"
)

func main() {
	// 创建消费者
	consumer, err := kafka.NewConsumer([]string{"localhost:9092"}, "my_group", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer consumer.Close()
	
	// 订阅主题
	consumer.Subscribe("my_topic", func(data []byte) {
		fmt.Println("Received:", string(data))
	})
	
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
	
	"github.com/oldfritter/tangram/lib/kafka"
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
	consumer, _ := kafka.NewConsumer([]string{"localhost:9092"}, "demo_group", nil)
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
	producer, _ := kafka.NewProducer([]string{"localhost:9092"}, nil)
	defer producer.Close()
	
	for i := 0; i < 5; i++ {
		msg := Message{
			Type:      "test",
			Content:   fmt.Sprintf("Message %d", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := producer.Publish(ctx, "demo_topic", "", msg); err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("Published message %d\n", i+1)
		time.Sleep(time.Second)
	}
}
```

## API 参考

### Producer

| 方法 | 说明 |
|------|------|
| `NewProducer(addrs, config)` | 创建生产者实例 |
| `Publish(ctx, topic, key, data)` | 发布消息到指定主题 |
| `PublishAsync(topic, key, data)` | 异步发布消息 |
| `Close()` | 关闭连接 |

### Consumer

| 方法 | 说明 |
|------|------|
| `NewConsumer(addrs, groupID, config)` | 创建消费者实例 |
| `Subscribe(topic, handler)` | 订阅单个主题 |
| `SubscribeMany(topics, handler)` | 订阅多个主题 |
| `Unsubscribe(topic)` | 取消订阅 |
| `Close()` | 关闭连接 |

## 配置说明

### 生产者配置

```go
config := sarama.NewConfig()
config.Producer.RequiredAcks = sarama.WaitForAll      // 等待所有副本确认
config.Producer.Retry.Max = 3                         // 最大重试次数
config.Producer.Return.Successes = true               // 返回成功消息
config.Producer.Compression = sarama.CompressionGZIP // 压缩方式
```

### 消费者配置

```go
config := sarama.NewConfig()
config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
config.Consumer.Offsets.Initial = sarama.OffsetNewest  // 从最新消息开始消费
```

## 注意事项

1. **Broker 地址**：使用 `host:port` 格式，如 `localhost:9092`
2. **消费者组**：同一消费者组内的消费者共同消费主题消息
3. **消息key**：用于分区器决定消息发送到哪个分区
4. **线程安全**：Producer 和 Consumer 都支持并发使用

## 与其他 MQ 对比

详见 `lib/redis/README.md` 中的对比表。
