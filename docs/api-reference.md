# API Reference

Full API reference for the Apinator Go SDK.

## Client

### `NewClient`

```go
func NewClient(appID, key, secret, cluster string, opts ...Option) *Client
```

Creates a new Apinator API client. The `cluster` parameter identifies the data plane region (e.g. `"eu"`, `"us"`). The HTTP base URL is derived as `https://ws-{cluster}.apinator.io`.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `appID` | `string` | Your Apinator application ID |
| `key` | `string` | Your API key |
| `secret` | `string` | Your API secret |
| `cluster` | `string` | Data plane cluster (e.g. `"eu"`, `"us"`) |
| `opts` | `...Option` | Optional configuration functions |

**Returns:** `*Client`

---

### `Client.Trigger`

```go
func (c *Client) Trigger(ctx context.Context, params TriggerParams) error
```

Sends an event to one or more channels. Either `Channel` or `Channels` must be provided (not both).

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `ctx` | `context.Context` | Request context |
| `params` | `TriggerParams` | Event parameters |

**Returns:** `error` — `nil` on success, or one of `*ApiError`, `*AuthenticationError`, `*ValidationError`.

---

### `Client.AuthenticateChannel`

```go
func (c *Client) AuthenticateChannel(socketID, channelName string, channelData *string) ChannelAuthResponse
```

Generates an authentication response for a channel subscription. For presence channels, `channelData` should be a JSON string containing `user_id` and optional `user_info`.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `socketID` | `string` | The socket ID from the client |
| `channelName` | `string` | The channel name being subscribed to |
| `channelData` | `*string` | Optional JSON channel data (required for presence channels) |

**Returns:** `ChannelAuthResponse`

---

### `Client.GetChannels`

```go
func (c *Client) GetChannels(ctx context.Context, prefix string) ([]ChannelInfo, error)
```

Retrieves a list of active channels, optionally filtered by prefix.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `ctx` | `context.Context` | Request context |
| `prefix` | `string` | Channel name prefix filter (e.g. `"presence-"`). Pass `""` for no filter. |

**Returns:** `[]ChannelInfo`, `error`

---

### `Client.GetChannel`

```go
func (c *Client) GetChannel(ctx context.Context, channelName string) (*ChannelInfo, error)
```

Retrieves information about a specific channel. Returns `nil, nil` if the channel does not exist.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `ctx` | `context.Context` | Request context |
| `channelName` | `string` | The channel name to query |

**Returns:** `*ChannelInfo`, `error`

---

### `Client.VerifyWebhook`

```go
func (c *Client) VerifyWebhook(headers http.Header, body []byte, maxAge time.Duration) bool
```

Verifies a webhook signature using the client's secret. Convenience method that delegates to the standalone `VerifyWebhook` function.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `headers` | `http.Header` | The HTTP headers from the webhook request |
| `body` | `[]byte` | The raw request body |
| `maxAge` | `time.Duration` | Maximum acceptable age of the webhook. Pass `0` to skip timestamp check. |

**Returns:** `bool` — `true` if the signature is valid.

---

## Standalone Functions

These functions can be used independently without creating a `Client` instance.

### `SignRequest`

```go
func SignRequest(secret, method, path string, body []byte, timestamp int64) string
```

Creates an HMAC-SHA256 signature for an API request. The signature string is computed as `"{timestamp}\n{method}\n{path}\n{bodyMD5}"` where `bodyMD5` is `hex(md5(body))` if body is non-empty, or `""` if empty.

**Returns:** Hex-encoded HMAC-SHA256 signature.

---

### `SignChannel`

```go
func SignChannel(secret, socketID, channelName string, channelData *string) string
```

Creates an auth signature for channel subscriptions. The signature string is `"{socketID}:{channelName}"`, with `":{channelData}"` appended if `channelData` is non-nil.

**Returns:** Hex-encoded HMAC-SHA256 signature.

---

### `AuthenticateChannel`

```go
func AuthenticateChannel(secret, key, socketID, channelName string, channelData *string) ChannelAuthResponse
```

Returns the full auth response for a channel subscription, with the `Auth` field formatted as `"{key}:{signature}"`.

**Returns:** `ChannelAuthResponse`

---

### `VerifyWebhook`

```go
func VerifyWebhook(secret string, headers http.Header, body []byte, maxAge time.Duration) bool
```

Verifies a webhook signature. Checks the `X-Realtime-Signature` header (strips `sha256=` prefix if present). If `maxAge > 0`, verifies timestamp freshness using the `X-Realtime-Timestamp` header.

**Returns:** `bool` — `true` if the signature is valid.

---

### `SignWebhookPayload`

```go
func SignWebhookPayload(secret, timestamp string, payload []byte) string
```

Signs a webhook payload. The input is `"{timestamp}.{payload}"`.

**Returns:** Hex-encoded HMAC-SHA256 signature.

---

## Types

### `TriggerParams`

```go
type TriggerParams struct {
    Name     string   `json:"name"`
    Channel  string   `json:"channel,omitempty"`
    Channels []string `json:"channels,omitempty"`
    Data     string   `json:"data"`
    SocketID string   `json:"socket_id,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Event name |
| `Channel` | `string` | Single channel name (mutually exclusive with `Channels`) |
| `Channels` | `[]string` | Multiple channel names (mutually exclusive with `Channel`) |
| `Data` | `string` | Event data as a JSON string |
| `SocketID` | `string` | Optional socket ID to exclude from receiving the event |

---

### `ChannelAuthResponse`

```go
type ChannelAuthResponse struct {
    Auth        string  `json:"auth"`
    ChannelData *string `json:"channel_data,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Auth` | `string` | Authentication string in the format `"{key}:{signature}"` |
| `ChannelData` | `*string` | Channel data for presence channels (JSON string) |

---

### `ChannelInfo`

```go
type ChannelInfo struct {
    Name              string `json:"name"`
    SubscriptionCount int    `json:"subscription_count"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Channel name |
| `SubscriptionCount` | `int` | Number of active subscriptions |

---

### `Option`

```go
type Option func(*Client)
```

Functional option for configuring the `Client`.

#### `WithHTTPClient`

```go
func WithHTTPClient(httpClient *http.Client) Option
```

Sets a custom `*http.Client` for the client. The default client has a 30-second timeout.

---

## Error Types

All error types implement the `error` interface and include `Message`, `Status`, and `Body` fields.

### `ApiError`

```go
type ApiError struct {
    Message string
    Status  int
    Body    string
}
```

General API error for non-specific HTTP error status codes (e.g. 500).

---

### `AuthenticationError`

```go
type AuthenticationError struct {
    Message string
    Status  int
    Body    string
}
```

Returned for HTTP 401 and 403 responses. Indicates invalid credentials or insufficient permissions.

---

### `ValidationError`

```go
type ValidationError struct {
    Message string
    Status  int
    Body    string
}
```

Returned for HTTP 400 and 422 responses. Indicates invalid request parameters.

---

## Error Handling Example

```go
err := client.Trigger(ctx, params)
if err != nil {
    switch e := err.(type) {
    case *apinator.AuthenticationError:
        log.Fatalf("Auth failed: %s (status %d)", e.Message, e.Status)
    case *apinator.ValidationError:
        log.Fatalf("Validation error: %s (status %d)", e.Message, e.Status)
    case *apinator.ApiError:
        log.Fatalf("API error: %s (status %d)", e.Message, e.Status)
    default:
        log.Fatalf("Unexpected error: %v", err)
    }
}
```
