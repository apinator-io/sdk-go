# Apinator Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/apinator-io/sdk-go.svg)](https://pkg.go.dev/github.com/apinator-io/sdk-go)
[![CI](https://github.com/apinator-io/sdk-go/actions/workflows/test.yml/badge.svg)](https://github.com/apinator-io/sdk-go/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Go server SDK for Apinator — trigger real-time events, authenticate channels, and verify webhooks.

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
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		socketID := r.FormValue("socket_id")
		channelName := r.FormValue("channel_name")

		auth := client.AuthenticateChannel(socketID, channelName, nil)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(auth)
	})

	http.ListenAndServe(":8080", nil)
}
```

For presence channels, include `channel_data`:

```go
channelData := `{"user_id": "123", "user_info": {"name": "Alice"}}`
auth := client.AuthenticateChannel(socketID, channelName, &channelData)
```

## Webhook Verification

Verify that incoming webhooks are authentic:

```go
package main

import (
	"io"
	"net/http"
	"time"

	apinator "github.com/apinator-io/sdk-go"
)

func main() {
	client := apinator.NewClient("app-id", "key", "secret", "eu")

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if !client.VerifyWebhook(r.Header, body, 5*time.Minute) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// Process the webhook payload...
		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(":8080", nil)
}
```

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
