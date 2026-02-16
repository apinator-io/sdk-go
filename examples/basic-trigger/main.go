package main

import (
	"context"
	"log"

	apinator "github.com/apinator-io/sdk-go"
)

func main() {
	client := apinator.NewClient(
		"your-app-id",
		"your-key",
		"your-secret",
		"eu",
	)

	// Trigger an event on a single channel
	err := client.Trigger(context.Background(), apinator.TriggerParams{
		Name:    "new-message",
		Channel: "chat-room",
		Data:    `{"user": "alice", "text": "Hello, world!"}`,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Event triggered successfully")

	// Trigger an event on multiple channels
	err = client.Trigger(context.Background(), apinator.TriggerParams{
		Name:     "price-update",
		Channels: []string{"stocks-AAPL", "stocks-GOOG"},
		Data:     `{"price": 150.25}`,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Multi-channel event triggered successfully")
}
