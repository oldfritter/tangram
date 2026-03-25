# Tangram - Distributed Service Integration Framework

> [中文版](./README_CN.MD)

Tangram is a framework for integrating distributed services across the network into a production system. It provides unified message queue communication between service nodes.

## Overview

Tangram enables seamless communication between microservices by providing a unified abstraction layer over various message queues (MQ). Services can communicate regardless of their physical location through a publish-subscribe pattern.

## Architecture

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

## Workflow

1. **User** visits WEB interface
2. **WEB** sends data query request to API node
3. **API Node** wraps request data and sends to other nodes via MQ
4. **Worker Nodes** listen to MQ, process data, and send results back via MQ
5. **API Node** returns data from MQ to the WEB interface

## Features

- **Unified MQ Interface**: Switch between Redis, Kafka, and RabbitMQ without changing business code
- **Multiple MQ Support**: Built-in support for Redis Pub/Sub, Kafka Producer/Consumer, and RabbitMQ
- **Easy Configuration**: Configure MQ settings via YAML file
- **Production Ready**: Designed for building distributed production systems
- **Language Independent**: Can be integrated with services written in any language

## Quick Start

### 1. Clone the project

```bash
git clone git@github.com:oldfritter/tangram.git
cd tangram
```

### 2. Configure MQ

Edit `example/config/app.yml`:

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

### 3. Run Examples

```bash
# Redis example
cd example/redis && go run main.go

# Kafka example
cd example/kafka && go run main.go

# RabbitMQ example
cd example/rabbitmq && go run main.go
```

## Project Structure

```
tangram/
├── mq.go                 # Unified MQ interface
├── MQ.md                 # MQ usage guide (English)
├── MQ_CN.md              # MQ usage guide (Chinese)
├── README.md             # This file
├── README_CN.MD          # Chinese version
├── config/
│   └── app.yml           # Configuration file
├── example/              # Usage examples
│   ├── config/          # Example config
│   ├── redis/           # Redis example
│   ├── kafka/           # Kafka example
│   └── rabbitmq/        # RabbitMQ example
└── lib/                  # MQ implementations
    ├── redis/           # Redis Pub/Sub
    ├── kafka/           # Kafka Producer/Consumer
    └── rabbitmq/        # RabbitMQ Queue/Exchange
```

## Supported Message Queues

| MQ Type | Implementation | Persistence | Consumer Groups | Best For |
|---------|---------------|-------------|-----------------|----------|
| Redis | Pub/Sub | ❌ | ❌ | Real-time messaging, low latency |
| Kafka | Producer/Consumer | ✅ | ✅ | High throughput, log aggregation |
| RabbitMQ | Queue/Exchange | ✅ | ✅ | Flexible routing, complex workflows |

## API Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/oldfritter/tangram"
)

func main() {
	// Load configuration
	cfg, err := mq.LoadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create MQ instance (automatically selects based on config)
	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	ctx := context.Background()
	topic := "my_topic"

	// Subscribe to messages
	mqInstance.Subscribe(topic, func(data []byte) {
		fmt.Println("Received:", string(data))
	})

	// Publish messages
	for i := 0; i < 5; i++ {
		msg := map[string]interface{}{
			"id":      i + 1,
			"message": "Hello from Tangram!",
		}
		if err := mqInstance.Publish(ctx, topic, msg); err != nil {
			log.Println("Error:", err)
		}
		time.Sleep(time.Second)
	}
}
```

## Switching MQ Implementation

To switch between message queue implementations, simply change the `mq.type` value in `example/config/app.yml`:

```yaml
# Use Redis
mq:
  type: "redis"

# Use Kafka
mq:
  type: "kafka"

# Use RabbitMQ
mq:
  type: "rabbitmq"
```

No code changes required!

## Requirements

- Go 1.19+
- Redis / Kafka / RabbitMQ (depending on which MQ you use)

## Documentation

- [MQ Guide (English)](./MQ.md)
- [MQ Guide (Chinese)](./MQ_CN.md)

## License

MIT License - see [LICENSE](./LICENSE) file for details.
