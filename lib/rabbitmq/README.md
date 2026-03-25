# RabbitMQ 使用指南

本模块提供基于 RabbitMQ 的消息通讯功能，支持队列和交换机两种模式。

## 快速开始

### 1. 安装依赖

```bash
go get github.com/rabbitmq/amqp091-go
```

### 2. 基本用法

#### 发布消息到队列

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/oldfritter/tangram/lib/rabbitmq"
)

func main() {
	ctx := context.Background()
	
	// 创建发布者
	pub, err := rabbitmq.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer pub.Close()
	
	// 发布消息到队列
	err = pub.Publish(ctx, "my_queue", map[string]interface{}{
		"type": "greeting",
		"msg":  "Hello!",
	})
	if err != nil {
		fmt.Println("Publish error:", err)
	}
}
```

#### 订阅队列消息

```go
package main

import (
	"fmt"
	"time"
	
	"github.com/oldfritter/tangram/lib/rabbitmq"
)

func main() {
	// 创建订阅者
	sub, err := rabbitmq.NewSubscriber("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer sub.Close()
	
	// 订阅队列
	err = sub.Subscribe("my_queue", func(data []byte) {
		fmt.Println("Received:", string(data))
	})
	if err != nil {
		fmt.Println("Subscribe error:", err)
	}
	
	time.Sleep(time.Second * 10)
}
```

### 3. 使用交换机（推荐）

RabbitMQ 支持四种交换机类型：direct、topic、headers、fanout。以下是 topic 类型的示例：

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"github.com/oldfritter/tangram/lib/rabbitmq"
)

func main() {
	ctx := context.Background()
	
	// 创建发布者
	pub, err := rabbitmq.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer pub.Close()
	
	// 发布到交换机
	for i := 0; i < 5; i++ {
		err := pub.PublishToExchange(ctx, "my_exchange", "order.created", map[string]interface{}{
			"order_id":  i + 1,
			"amount":    100.0 * float64(i+1),
		})
		if err != nil {
			log.Println("Publish error:", err)
		}
		time.Sleep(time.Second)
	}
}

// 另一个进程订阅
func consume() {
	sub, err := rabbitmq.NewSubscriber("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Close()
	
	// 绑定到交换机
	err = sub.SubscribeToExchange("my_exchange", "order_queue", "order.*", func(data []byte) {
		fmt.Println("Received:", string(data))
	})
	if err != nil {
		log.Fatal(err)
	}
	
	select {}
}
```

### 4. 完整示例

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	
	"github.com/oldfritter/tangram/lib/rabbitmq"
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
	consumer, err := rabbitmq.NewSubscriber("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()
	
	err = consumer.Subscribe("demo_queue", func(data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("Received: [Type=%s] %s\n", msg.Type, msg.Content)
	})
	if err != nil {
		log.Fatal(err)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	// 启动生产者
	producer, err := rabbitmq.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()
	
	for i := 0; i < 5; i++ {
		msg := Message{
			Type:      "test",
			Content:   fmt.Sprintf("Message %d", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := producer.Publish(ctx, "demo_queue", msg); err != nil {
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
| `NewPublisher(addr)` | 创建发布者实例 |
| `Publish(ctx, queue, data)` | 发布消息到队列 |
| `PublishToExchange(ctx, exchange, routingKey, data)` | 发布消息到交换机 |
| `Close()` | 关闭连接 |

### Subscriber

| 方法 | 说明 |
|------|------|
| `NewSubscriber(addr)` | 创建订阅者实例 |
| `Subscribe(queue, handler)` | 订阅队列 |
| `SubscribeToExchange(exchange, queue, routingKey, handler)` | 订阅交换机 |
| `SubscribeMany(queues, handler)` | 订阅多个队列 |
| `Unsubscribe(queue)` | 取消订阅 |
| `Close()` | 关闭连接 |

## 连接地址格式

```
amqp://username:password@host:port/
```

默认账号：`guest:guest`（仅限 localhost）

## 交换机类型

| 类型 | 说明 | 路由规则 |
|------|------|----------|
| direct | 精确匹配 | routing key 完全匹配 |
| topic | 模糊匹配 | 支持 `*` 和 `#` 通配符 |
| headers | 消息头匹配 | 根据消息头属性路由 |
| fanout | 广播 | 发送到所有绑定的队列 |

### Topic 交换机通配符

- `*` 匹配一个单词
- `#` 匹配零个或多个单词

示例：
- `order.*` 匹配 `order.created`、`order.updated`
- `order.#` 匹配 `order`、`order.created`、`order.created.processed`

## 注意事项

1. **消息持久化**：队列和消息都设置为持久化
2. **自动确认**：默认关闭，需要手动 ACK
3. **消费者预取**：可设置 `channel.Qos()` 限制并发消费数量
4. **线程安全**：建议每个 goroutine 使用独立的 channel

## 与其他 MQ 对比

详见 `lib/redis/README.md` 中的对比表。
