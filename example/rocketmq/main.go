package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mq "github.com/oldfritter/tangram"
)

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	cfg, err := mq.LoadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	cfg.Type = "rocketmq"
	cfg.RocketMQ.NameServer = "localhost:9876"
	cfg.RocketMQ.GroupID = "demo_group"

	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	fmt.Printf("Using MQ Type: %s (message type: %s)\n", mqInstance.GetType(), mqInstance.GetMsgType())

	ctx := context.Background()
	topic := "demo_topic"

	mqInstance.Subscribe(topic, func(data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("[%s] Received: [Type=%s] %s\n", mqInstance.GetType(), msg.Type, msg.Content)
	})

	time.Sleep(500 * time.Millisecond)

	for i := 0; i < 5; i++ {
		msg := Message{
			Type:      "greeting",
			Content:   fmt.Sprintf("Hello from RocketMQ! (#%d)", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := mqInstance.Publish(ctx, topic, msg); err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("[RocketMQ] Published message %d\n", i+1)
		time.Sleep(time.Second)
	}

	fmt.Println("Done!")
}
