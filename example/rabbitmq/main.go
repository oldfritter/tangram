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

	cfg.Type = "rabbitmq"
	cfg.RabbitMQ.Addr = "amqp://guest:guest@localhost:5672/"

	mqInstance, err := mq.NewMQ(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer mqInstance.Close()

	fmt.Printf("Using MQ Type: %s (message type: %s)\n", mqInstance.GetType(), mqInstance.GetMsgType())

	ctx := context.Background()
	queue := "demo_queue"

	mqInstance.Subscribe(queue, func(data []byte) {
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
			Content:   fmt.Sprintf("Hello from RabbitMQ! (#%d)", i+1),
			Timestamp: time.Now().Unix(),
		}
		if err := mqInstance.Publish(ctx, queue, msg); err != nil {
			log.Println("Publish error:", err)
		}
		fmt.Printf("[RabbitMQ] Published message %d\n", i+1)
		time.Sleep(time.Second)
	}

	fmt.Println("Done!")
}
