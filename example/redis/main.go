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
	// 加载配置
	cfg, err := mq.LoadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 强制使用 Redis
	cfg.Type = "redis"
	cfg.Redis.Addr = "localhost:6379"
	cfg.Redis.Password = ""
	cfg.Redis.DB = 0

	// 创建 MQ 实例
	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	fmt.Printf("Using MQ Type: %s (message type: %s)\n", mqInstance.GetType(), mqInstance.GetMsgType())

	ctx := context.Background()
	channel := "demo_channel"

	// 启动消费者
	mqInstance.Subscribe(channel, func(data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("[%s] Received: [Type=%s] %s\n", mqInstance.GetType(), msg.Type, msg.Content)
	})

	time.Sleep(500 * time.Millisecond)

	// 启动生产者
	for i := 0; i < 5; i++ {
		msg := Message{
			Type:      "greeting",
			Content:   fmt.Sprintf("Hello from Redis Pub/Sub! (#%d)", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := mqInstance.Publish(ctx, channel, msg); err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("[Redis] Published message %d\n", i+1)
		time.Sleep(time.Second)
	}

	fmt.Println("Done!")
}
