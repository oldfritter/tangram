# Redis Pub/Sub 使用指南

本模块提供基于 Redis Pub/Sub 的消息通讯功能，支持发布/订阅模式。

## 快速开始

### 1. 安装依赖

确保项目中已包含 redis 客户端：

```bash
go get github.com/redis/go-redis/v9
```

### 2. 基本用法

#### 发布消息

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/oldfritter/tangram/lib/redis"
)

func main() {
	ctx := context.Background()
	
	// 创建发布者
	pub := redis.NewPublisher("localhost:6379", "", 0)
	defer pub.Close()
	
	// 发布消息（支持 string、[]byte、任意可序列化类型）
	err := pub.Publish(ctx, "my_channel", map[string]interface{}{
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
	
	"github.com/oldfritter/tangram/lib/redis"
)

func main() {
	// 创建订阅者
	sub := redis.NewSubscriber("localhost:6379", "", 0)
	defer sub.Close()
	
	// 订阅单个频道
	sub.Subscribe("my_channel", func(data []byte) {
		fmt.Println("Received:", string(data))
	})
	
	// 保持程序运行
	time.Sleep(time.Second * 10)
}
```

#### 订阅多个频道

```go
// 订阅多个频道，共用一个处理函数
sub.SubscribeMany([]string{"channel1", "channel2", "channel3"}, func(data []byte) {
	fmt.Println("Received:", string(data))
})
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
	
	"github.com/oldfritter/tangram/lib/redis"
)

// 自定义消息结构
type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Timestamp int64 `json:"timestamp"`
}

func main() {
	ctx := context.Background()
	
	// 启动订阅者（消费者）
	go runSubscriber()
	
	// 等待订阅者就绪
	time.Sleep(500 * time.Millisecond)
	
	// 启动发布者（生产者）
	runPublisher(ctx)
}

func runPublisher(ctx context.Context) {
	pub := redis.NewPublisher("localhost:6379", "", 0)
	defer pub.Close()
	
	msg := Message{
		Type:    "test",
		Content: "Hello from Redis Pub/Sub!",
		Timestamp: time.Now().Unix(),
	}
	
	for i := 0; i < 5; i++ {
		err := pub.Publish(ctx, "demo_channel", msg)
		if err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("Published message %d\n", i+1)
		time.Sleep(time.Second)
	}
}

func runSubscriber() {
	sub := redis.NewSubscriber("localhost:6379", "", 0)
	defer sub.Close()
	
	sub.Subscribe("demo_channel", func(data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("Received: [Type=%s] %s\n", msg.Type, msg.Content)
	})
	
	// 阻塞等待
	select {}
}
```

## API 参考

### Publisher

| 方法 | 说明 |
|------|------|
| `NewPublisher(addr, password, db)` | 创建发布者实例 |
| `Publish(ctx, channel, data)` | 发布消息到指定频道 |
| `Ping(ctx)` | 检查 Redis 连接 |
| `Close()` | 关闭连接 |

### Subscriber

| 方法 | 说明 |
|------|------|
| `NewSubscriber(addr, password, db)` | 创建订阅者实例 |
| `Subscribe(channel, handler)` | 订阅单个频道 |
| `SubscribeMany(channels, handler)` | 订阅多个频道 |
| `Unsubscribe(channel)` | 取消订阅 |
| `Ping(ctx)` | 检查 Redis 连接 |
| `Close()` | 关闭连接 |

## 注意事项

1. **连接参数**：`addr` 格式为 `host:port`，如 `localhost:6379`
2. **密码认证**：如无需密码请传空字符串 `""`
3. **数据库编号**：Redis 默认 DB 为 0
4. **线程安全**：`Publisher` 支持并发发布
5. **消息格式**：发布 `struct`/`map` 时会自动 JSON 序列化，也可直接发送 `string` 或 `[]byte`

## 与其他 MQ 对比

| 特性 | Redis Pub/Sub | Kafka | RabbitMQ |
|------|---------------|-------|----------|
| 延迟 | 低 | 中 | 低 |
| 持久化 | ❌ | ✅ | ✅ |
| 消息确认 | ❌ | ✅ | ✅ |
| 消费者组 | ❌ | ✅ | ✅ |
| 复杂度 | 简单 | 中等 | 较复杂 |

适用场景：实时通讯、简单队列、对消息持久化无要求的场景。
