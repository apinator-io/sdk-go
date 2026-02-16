# Quick Start

Get up and running with the Apinator Go SDK in minutes.

## Installation

```bash
go get github.com/apinator-io/sdk-go
```

Requires Go 1.21 or later.

## Create a Client

```go
import apinator "github.com/apinator-io/sdk-go"

client := apinator.NewClient(
    "your-app-id",
    "your-key",
    "your-secret",
    "eu", // cluster
)
```

The `cluster` parameter determines which data plane region to connect to. The SDK constructs the base URL as `https://ws-{cluster}.apinator.io`.

## Trigger an Event

Send an event to a channel:

```go
err := client.Trigger(context.Background(), apinator.TriggerParams{
    Name:    "new-message",
    Channel: "chat-room",
    Data:    `{"text": "Hello, world!"}`,
})
if err != nil {
    log.Fatal(err)
}
```

Send to multiple channels at once:

```go
err := client.Trigger(context.Background(), apinator.TriggerParams{
    Name:     "price-update",
    Channels: []string{"stocks-AAPL", "stocks-GOOG"},
    Data:     `{"price": 150.25}`,
})
```

## Channel Authentication

When a client subscribes to a private or presence channel, your server must authenticate the request. Set up an HTTP endpoint:

```go
http.HandleFunc("/realtime/auth", func(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()

    socketID := r.FormValue("socket_id")
    channelName := r.FormValue("channel_name")

    // For private channels
    auth := client.AuthenticateChannel(socketID, channelName, nil)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(auth)
})
```

For presence channels, include channel data with the user's identity:

```go
channelData := `{"user_id": "user-123", "user_info": {"name": "Alice"}}`
auth := client.AuthenticateChannel(socketID, channelName, &channelData)
```

## Webhook Verification

Verify that incoming webhooks are sent by Apinator:

```go
http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    if !client.VerifyWebhook(r.Header, body, 5*time.Minute) {
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }

    // Webhook is authentic — process the payload
    w.WriteHeader(http.StatusOK)
})
```

The `maxAge` parameter (e.g. `5*time.Minute`) rejects webhooks with timestamps older than the specified duration. Pass `0` to skip the timestamp check.

## Next Steps

- [API Reference](api-reference.md) for the full API surface
- [Examples](../examples/) for runnable code samples
