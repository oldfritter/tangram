# Unified Message Queue (MQ) Guide

> [中文版](./MQ_CN.md)

This module provides a unified message queue interface that automatically selects the underlying implementation (Redis, Kafka, or RabbitMQ) based on the configuration file.

## Quick Start

### 1. Configure MQ

Configure the MQ type in `example/config/app.yml`:

```yaml
mq:
  type: "kafka"  # Options: redis, kafka, rabbitmq

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
```

### 2. Usage Example

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
	// Auto-load config from example/config/app.yml
	cfg, err := mq.LoadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create MQ instance based on config
	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	ctx := context.Background()
	topic := "demo_topic"

	// Start consumer
	mqInstance.Subscribe(topic, func(data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("Received: [Type=%s] %s\n", msg.Type, msg.Content)
	})

	time.Sleep(500 * time.Millisecond)

	// Start producer
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

## Configuration Reference

| Field | Description | Example |
|-------|-------------|---------|
| `mq.type` | MQ type | `kafka`, `redis`, `rabbitmq` |
| `mq.redis.addr` | Redis address | `localhost:6379` |
| `mq.redis.password` | Redis password | `""` |
| `mq.redis.db` | Redis database number | `0` |
| `mq.kafka.addrs` | Kafka broker addresses | `["localhost:9092"]` |
| `mq.kafka.groupId` | Consumer group ID | `my_group` |
| `mq.rabbitmq.addr` | RabbitMQ address | `amqp://guest:guest@localhost:5672/` |

## API Reference

| Method | Description |
|--------|-------------|
| `LoadDefaultConfig()` | Auto-load `example/config/app.yml` |
| `LoadConfigFromYAML(path)` | Load YAML config from specified path |
| `LoadConfigFromYAMLString(yaml)` | Load config from YAML string |
| `NewMQ(cfg)` | Create MQ instance from config |
| `Publish(ctx, topic, data)` | Publish message |
| `Subscribe(topic, handler)` | Subscribe to messages |
| `SubscribeMany(topics, handler)` | Subscribe to multiple topics |
| `Unsubscribe(topic)` | Unsubscribe |
| `Close()` | Close connection |
| `GetType()` | Get current MQ type |
| `GetMsgType()` | Get message type name |

## Switching MQ Implementation

Simply modify the `mq.type` field in `example/config/app.yml`, no business code changes needed:

| type value | Implementation |
|-----------|----------------|
| `"redis"` | Redis Pub/Sub |
| `"kafka"` | Kafka Producer/Consumer |
| `"rabbitmq"` | RabbitMQ Queue |

## Topic Parameter Meaning

| MQ Type | Topic Parameter Meaning |
|---------|------------------------|
| Redis | channel name |
| Kafka | topic name |
| RabbitMQ | queue name |

## Project Structure

```
your-project/
├── example/
│   ├── config/
│   │   └── app.yml     # MQ config file
│   ├── redis/
│   ├── kafka/
│   └── rabbitmq/
├── mq.go               # Unified MQ interface
└── main.go             # Business code
```
