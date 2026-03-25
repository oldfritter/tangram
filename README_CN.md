# Tangram - 分布式服务集成框架

> [English](./README.md)

Tangram 是一个用于将分布在网络各处的服务整合成生产系统的框架。它在服务节点之间提供统一的消息队列通讯层。

## 概述

Tangram 通过在各种消息队列（MQ）上提供统一的抽象层，实现微服务之间的无缝通讯。服务可以通过发布-订阅模式无论其物理位置如何都能进行通讯。

## 架构

```
┌─────────┐      ┌─────────┐      ┌─────────┐      ┌─────────┐
│   WEB   │─────▶│   API   │─────▶│   MQ   │◀─────▶│  Node  │
│   UI    │      │  Node   │      │ Broker │       │   1    │
└─────────┘      └─────────┘      └─────────┘      └─────────┘
                                            │
                                            ▼
                                       ┌─────────┐
                                       │  Node   │
                                       │   2     │
                                       └─────────┘
```

## 工作流程

1. **用户** 访问 WEB 界面
2. **WEB** 发送数据查询请求到 API 节点
3. **API 节点** 包装请求数据，通过 MQ 发送到其他节点
4. **工作节点** 监听 MQ，收到数据后进行处理，并通过 MQ 将结果发送回 API 节点
5. **API 节点** 将从 MQ 收到的数据返回给 WEB 界面

## 功能特性

- **统一 MQ 接口**：无需修改业务代码即可在 Redis、Kafka 和 RabbitMQ 之间切换
- **多种 MQ 支持**：内置支持 Redis Pub/Sub、Kafka Producer/Consumer 和 RabbitMQ
- **简易配置**：通过 YAML 文件配置 MQ 设置
- **生产就绪**：专为构建分布式生产系统设计
- **语言无关**：可与任何语言编写的服务集成

## 快速开始

### 1. 克隆项目

```bash
git clone git@github.com:oldfritter/tangram.git
cd tangram
```

### 2. 配置 MQ

编辑 `example/config/app.yml`：

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

### 3. 运行示例

```bash
# Redis 示例
cd example/redis && go run main.go

# Kafka 示例
cd example/kafka && go run main.go

# RabbitMQ 示例
cd example/rabbitmq && go run main.go

# RocketMQ 示例
cd example/rocketmq && go run main.go
```

## 项目结构

```
tangram/
├── mq.go                 # 统一 MQ 接口
├── MQ.md                 # MQ 使用指南（英文）
├── MQ_CN.md              # MQ 使用指南（中文）
├── README.md             # 本文件
├── README_CN.md          # 中文版本
├── example/              # 使用示例
│   ├── config/           # 示例配置
│   ├── redis/            # Redis 示例
│   ├── kafka/            # Kafka 示例
│   ├── rabbitmq/         # RabbitMQ 示例
│   └── rocketmq/         # RocketMQ 示例
└── lib/                  # MQ 实现
    ├── redis/            # Redis Pub/Sub
    ├── kafka/            # Kafka Producer/Consumer
    ├── rabbitmq/         # RabbitMQ 队列/交换机
    └── rocketmq/        # RocketMQ Producer/Consumer
```

## 支持的消息队列

| MQ 类型 | 实现方式 | 持久化 | 消费者组 | 适用场景 |
|---------|----------|--------|----------|----------|
| Redis | Pub/Sub | ❌ | ❌ | 实时通讯、低延迟 |
| Kafka | Producer/Consumer | ✅ | ✅ | 高吞吐量、日志聚合 |
| RabbitMQ | 队列/交换机 | ✅ | ✅ | 灵活路由、复杂工作流 |
| RocketMQ | Producer/Consumer | ✅ | ✅ | 分布式事务、顺序消息 |

## API 使用

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mq "github.com/oldfritter/tangram"
)

func main() {
	// 加载配置
	cfg, err := mq.LoadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 创建 MQ 实例（根据配置自动选择）
	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	ctx := context.Background()
	topic := "my_topic"

	// 订阅消息
	mqInstance.Subscribe(topic, func(data []byte) {
		fmt.Println("收到:", string(data))
	})

	// 发布消息
	for i := 0; i < 5; i++ {
		msg := map[string]interface{}{
			"id":      i + 1,
			"message": "Hello from Tangram!",
		}
		if err := mqInstance.Publish(ctx, topic, msg); err != nil {
			log.Println("错误:", err)
		}
		time.Sleep(time.Second)
	}
}
```

## 切换 MQ 实现

只需修改 `example/config/app.yml` 中的 `mq.type` 值即可切换消息队列实现：

```yaml
# 使用 Redis
mq:
  type: "redis"

# 使用 Kafka
mq:
  type: "kafka"

# 使用 RabbitMQ
mq:
  type: "rabbitmq"
```

无需修改任何代码！

## 环境要求

- Go 1.19+
- Redis / Kafka / RabbitMQ（取决于使用的 MQ）

## 文档

- [MQ 使用指南（英文）](./MQ.md)
- [MQ 使用指南（中文）](./MQ_CN.md)

## 许可证

MIT License - 见 [LICENSE](./LICENSE) 文件了解详情。
