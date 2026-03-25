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

	// 强制使用 Kafka
	cfg.Type = "kafka"
	cfg.Kafka.Addrs = []string{"localhost:9092"}
	cfg.Kafka.GroupID = "demo_group"

	// 创建 MQ 实例
	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	fmt.Printf("Using MQ Type: %s (message type: %s)\n", mqInstance.GetType(), mqInstance.GetMsgType())

	ctx := context.Background()
	topic := "demo_topic"

	// 启动消费者
	mqInstance.Subscribe(topic, func(data []byte) {
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
			Content:   fmt.Sprintf("Hello from Kafka! (#%d)", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := mqInstance.Publish(ctx, topic, msg); err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("[Kafka] Published message %d\n", i+1)
		time.Sleep(time.Second)
	}

	fmt.Println("Done!")
}
