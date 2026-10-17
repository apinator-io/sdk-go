# Apinator Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/apinator-io/sdk-go.svg)](https://pkg.go.dev/github.com/apinator-io/sdk-go)
[![CI](https://github.com/apinator-io/sdk-go/actions/workflows/test.yml/badge.svg)](https://github.com/apinator-io/sdk-go/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Go server SDK for Apinator — trigger real-time events, authenticate channels.

## Features

- Trigger events to one or more channels
- Authenticate private and presence channel subscriptions
- Verify incoming webhook signatures
- Query channel state and subscription counts
- HMAC-SHA256 request signing
- Zero external dependencies (stdlib only)
- Go 1.21+

## Installation

```bash
go get github.com/apinator-io/sdk-go
```

## Quick Start

```go
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
		"eu", // cluster
	)

	err := client.Trigger(context.Background(), apinator.TriggerParams{
		Name:    "my-event",
		Channel: "my-channel",
		Data:    `{"message": "hello world"}`,
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## Cache Channels

Publish to a channel named `cache-`, `private-cache-` or `presence-cache-` and Apinator keeps
the last event for 30 minutes, so a client that subscribes after the publish still receives it.
Nothing else changes: trigger as usual, and authorize private and presence cache channels like
any other private or presence channel.

```go
err := client.Trigger(ctx, apinator.TriggerParams{
	Name:    "job-finished",
	Channel: "private-cache-job-" + jobID,
	Data:    `{"status": "done"}`,
})
```

## Channel Authentication

Authenticate private and presence channel subscriptions from your server:

```go
package main

import (
	"encoding/json"
	"net/http"

	apinator "github.com/apinator-io/sdk-go"
)

func main() {
	client := apinator.NewClient("app-id", "key", "secret", "eu")

	http.HandleFunc("/realtime/auth", func(w http.ResponseWriter, r *http.Request) {
		// The client SDKs POST a JSON body.
		var req struct {
			SocketID    string `json:"socket_id"`
			ChannelName string `json:"channel_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		auth := client.AuthenticateChannel(req.SocketID, req.ChannelName, nil)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(auth)
	})

	http.ListenAndServe(":8080", nil)
}
```

For presence channels, include `channel_data`:

```go
channelData := `{"user_id": "123", "user_info": {"name": "Alice"}}`
auth := client.AuthenticateChannel(req.SocketID, req.ChannelName, &channelData)
```

## Webhook Verification

Each webhook endpoint you create in the console has its own signing secret
(`whsec_...`). Verify deliveries with that secret, not the app's API secret, and
pass the raw request body exactly as received:

```go
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	apinator "github.com/apinator-io/sdk-go"
)

func main() {
	client := apinator.NewClient("app-id", "key", "secret", "eu")
	webhookSecret := os.Getenv("APINATOR_WEBHOOK_SECRET") // whsec_...

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if !client.VerifyWebhook(webhookSecret, r.Header, body, 5*time.Minute) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// {"id":"<event uuid>","event":"member_added","app_id":"...","channel":"presence-room",
		//  "user_id":"42","time_ms":1760000000000}
		var payload struct {
			ID      string `json:"id"`
			Event   string `json:"event"`
			Channel string `json:"channel"`
			UserID  string `json:"user_id"` // member_added / member_removed only
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Deliveries are at-least-once: skip payload.ID if already processed.

		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(":8080", nil)
}
```

Each delivery carries `X-Realtime-Signature` (`sha256=<hex>`),
`X-Realtime-Timestamp` (unix seconds), `X-Realtime-Event` and
`X-Realtime-Delivery-ID`. Events: `channel_occupied`, `channel_vacated`,
`member_added`, `member_removed`; the console's test button sends `webhook.test`.
With `maxAge > 0`, timestamps older than `maxAge` or more than 300 seconds in the
future are rejected.

## Channel Introspection

Query active channels and their subscription counts:

```go
// List all channels
channels, err := client.GetChannels(ctx, "")

// Filter by prefix
channels, err := client.GetChannels(ctx, "presence-")

// Get a specific channel
channel, err := client.GetChannel(ctx, "my-channel")
if channel != nil {
	fmt.Printf("Channel: %s, Subscribers: %d\n", channel.Name, channel.SubscriptionCount)
}
```

## Options

```go
// Use a custom HTTP client
client := apinator.NewClient("app-id", "key", "secret", "eu",
	apinator.WithHTTPClient(&http.Client{
		Timeout: 10 * time.Second,
	}),
)
```

## Links

- [API Reference](docs/api-reference.md)
- [Quick Start Guide](docs/quickstart.md)
- [Examples](examples/)
- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)
- [Apinator Documentation](https://docs.apinator.io)
- [pkg.go.dev](https://pkg.go.dev/github.com/apinator-io/sdk-go)
