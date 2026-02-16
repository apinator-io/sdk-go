package apinator

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("app-123", "key-456", "secret-789", "eu")

	if client.appID != "app-123" {
		t.Errorf("expected appID %q, got %q", "app-123", client.appID)
	}

	if client.key != "key-456" {
		t.Errorf("expected key %q, got %q", "key-456", client.key)
	}

	if client.secret != "secret-789" {
		t.Errorf("expected secret %q, got %q", "secret-789", client.secret)
	}

	if client.host != "https://ws-eu.apinator.io" {
		t.Errorf("expected default host %q, got %q", "https://ws-eu.apinator.io", client.host)
	}

	if client.http == nil {
		t.Error("expected http client to be initialized")
	}

	if client.http.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", client.http.Timeout)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 10 * time.Second}

	client := NewClient(
		"app-123",
		"key-456",
		"secret-789",
		"eu",
		WithHTTPClient(customHTTP),
	)

	if client.host != "https://ws-eu.apinator.io" {
		t.Errorf("expected host %q, got %q", "https://ws-eu.apinator.io", client.host)
	}

	if client.http != customHTTP {
		t.Error("expected custom http client")
	}
}

func TestNewClient_USCluster(t *testing.T) {
	client := NewClient("app-123", "key-456", "secret-789", "us")

	if client.host != "https://ws-us.apinator.io" {
		t.Errorf("expected host %q, got %q", "https://ws-us.apinator.io", client.host)
	}
}

func TestTrigger(t *testing.T) {
	// Track request details
	var receivedMethod string
	var receivedPath string
	var receivedBody []byte
	var receivedHeaders http.Header

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		receivedHeaders = r.Header.Clone()

		body, _ := io.ReadAll(r.Body)
		receivedBody = body

		// Verify auth headers are present
		if r.Header.Get("X-Realtime-Key") == "" {
			t.Error("missing X-Realtime-Key header")
		}
		if r.Header.Get("X-Realtime-Timestamp") == "" {
			t.Error("missing X-Realtime-Timestamp header")
		}
		if r.Header.Get("X-Realtime-Signature") == "" {
			t.Error("missing X-Realtime-Signature header")
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	// Create client
	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	// Trigger event
	params := TriggerParams{
		Name:    "test-event",
		Channel: "test-channel",
		Data:    `{"message":"hello"}`,
	}

	err := client.Trigger(context.Background(), params)
	if err != nil {
		t.Fatalf("trigger failed: %v", err)
	}

	// Verify request details
	if receivedMethod != "POST" {
		t.Errorf("expected method POST, got %s", receivedMethod)
	}

	expectedPath := "/apps/app-123/events"
	if receivedPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, receivedPath)
	}

	// Verify body
	var receivedParams TriggerParams
	if err := json.Unmarshal(receivedBody, &receivedParams); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if receivedParams.Name != params.Name {
		t.Errorf("expected name %q, got %q", params.Name, receivedParams.Name)
	}

	if receivedParams.Channel != params.Channel {
		t.Errorf("expected channel %q, got %q", params.Channel, receivedParams.Channel)
	}

	// Verify content type
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", receivedHeaders.Get("Content-Type"))
	}

	// Verify auth headers format
	if receivedHeaders.Get("X-Realtime-Key") != "key-456" {
		t.Errorf("expected key %q, got %q", "key-456", receivedHeaders.Get("X-Realtime-Key"))
	}

	timestamp := receivedHeaders.Get("X-Realtime-Timestamp")
	if _, err := strconv.ParseInt(timestamp, 10, 64); err != nil {
		t.Errorf("timestamp should be valid int64: %v", err)
	}

	signature := receivedHeaders.Get("X-Realtime-Signature")
	if len(signature) != 64 {
		t.Errorf("signature should be 64 hex chars, got %d", len(signature))
	}
}

func TestTrigger_MultipleChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var params TriggerParams
		json.Unmarshal(body, &params)

		// Verify channels array is present
		if len(params.Channels) != 2 {
			t.Errorf("expected 2 channels, got %d", len(params.Channels))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	params := TriggerParams{
		Name:     "test-event",
		Channels: []string{"channel-1", "channel-2"},
		Data:     `{"message":"hello"}`,
	}

	err := client.Trigger(context.Background(), params)
	if err != nil {
		t.Fatalf("trigger failed: %v", err)
	}
}

func TestTrigger_WithSocketID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var params TriggerParams
		json.Unmarshal(body, &params)

		if params.SocketID != "12345.67890" {
			t.Errorf("expected socketID %q, got %q", "12345.67890", params.SocketID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	params := TriggerParams{
		Name:     "test-event",
		Channel:  "test-channel",
		Data:     `{"message":"hello"}`,
		SocketID: "12345.67890",
	}

	err := client.Trigger(context.Background(), params)
	if err != nil {
		t.Fatalf("trigger failed: %v", err)
	}
}

func TestTrigger_AuthenticationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type":   "https://docs.apinator.io/problems/unauthorized",
			"title":  "Unauthorized",
			"status": 401,
			"detail": "Invalid credentials",
			"code":   "unauthorized",
		})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	params := TriggerParams{
		Name:    "test-event",
		Channel: "test-channel",
		Data:    `{"message":"hello"}`,
	}

	err := client.Trigger(context.Background(), params)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify error type
	authErr, ok := err.(*AuthenticationError)
	if !ok {
		t.Fatalf("expected AuthenticationError, got %T", err)
	}

	if authErr.Status != 401 {
		t.Errorf("expected status 401, got %d", authErr.Status)
	}

	if !strings.Contains(authErr.Message, "Invalid credentials") {
		t.Errorf("expected error message to contain 'Invalid credentials', got %q", authErr.Message)
	}
}

func TestTrigger_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type":   "https://docs.apinator.io/problems/bad_request",
			"title":  "Bad Request",
			"status": 400,
			"detail": "Invalid event name",
			"code":   "bad_request",
		})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	params := TriggerParams{
		Name:    "",
		Channel: "test-channel",
		Data:    `{"message":"hello"}`,
	}

	err := client.Trigger(context.Background(), params)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify error type
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if valErr.Status != 400 {
		t.Errorf("expected status 400, got %d", valErr.Status)
	}
}

func TestGetChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		expectedPath := "/apps/app-123/channels"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, r.URL.Path)
		}

		// Verify auth headers
		if r.Header.Get("X-Realtime-Key") == "" {
			t.Error("missing X-Realtime-Key header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"channels": []ChannelInfo{
				{Name: "channel-1", SubscriptionCount: 5},
				{Name: "channel-2", SubscriptionCount: 3},
			},
		})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	channels, err := client.GetChannels(context.Background(), "")
	if err != nil {
		t.Fatalf("GetChannels failed: %v", err)
	}

	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}

	if channels[0].Name != "channel-1" {
		t.Errorf("expected channel name %q, got %q", "channel-1", channels[0].Name)
	}

	if channels[0].SubscriptionCount != 5 {
		t.Errorf("expected subscription count 5, got %d", channels[0].SubscriptionCount)
	}
}

func TestGetChannels_WithPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := r.URL.Query().Get("filter_by_prefix")
		if prefix != "private-" {
			t.Errorf("expected prefix %q, got %q", "private-", prefix)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"channels": []ChannelInfo{
				{Name: "private-chat", SubscriptionCount: 2},
			},
		})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	channels, err := client.GetChannels(context.Background(), "private-")
	if err != nil {
		t.Fatalf("GetChannels failed: %v", err)
	}

	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}

	if channels[0].Name != "private-chat" {
		t.Errorf("expected channel name %q, got %q", "private-chat", channels[0].Name)
	}
}

func TestGetChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		expectedPath := "/apps/app-123/channels/test-channel"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChannelInfo{
			Name:              "test-channel",
			SubscriptionCount: 10,
		})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	channel, err := client.GetChannel(context.Background(), "test-channel")
	if err != nil {
		t.Fatalf("GetChannel failed: %v", err)
	}

	if channel == nil {
		t.Fatal("expected channel, got nil")
	}

	if channel.Name != "test-channel" {
		t.Errorf("expected channel name %q, got %q", "test-channel", channel.Name)
	}

	if channel.SubscriptionCount != 10 {
		t.Errorf("expected subscription count 10, got %d", channel.SubscriptionCount)
	}
}

func TestGetChannel_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type":   "https://docs.apinator.io/problems/not_found",
			"title":  "Not Found",
			"status": 404,
			"detail": "Channel not found",
			"code":   "not_found",
		})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	channel, err := client.GetChannel(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected no error for 404, got %v", err)
	}

	if channel != nil {
		t.Errorf("expected nil channel for 404, got %v", channel)
	}
}

func TestGetChannel_WithSpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note: url.PathEscape doesn't escape : because it's valid in paths
		// But it does escape spaces and other special chars
		if !strings.Contains(r.URL.Path, "/apps/app-123/channels/") {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChannelInfo{
			Name:              "private-user:123",
			SubscriptionCount: 1,
		})
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	// Test with colon (not escaped by PathEscape)
	channel, err := client.GetChannel(context.Background(), "private-user:123")
	if err != nil {
		t.Fatalf("GetChannel failed: %v", err)
	}

	if channel == nil {
		t.Fatal("expected channel, got nil")
	}

	if channel.Name != "private-user:123" {
		t.Errorf("expected channel name %q, got %q", "private-user:123", channel.Name)
	}
}

func TestAuthenticateChannel_Method(t *testing.T) {
	client := NewClient("app-123", "key-456", "secret-789", "eu")

	resp := client.AuthenticateChannel("12345.67890", "private-chat", nil)

	// Verify auth format
	if !strings.HasPrefix(resp.Auth, "key-456:") {
		t.Errorf("auth should start with key, got %q", resp.Auth)
	}

	if len(resp.Auth) < len("key-456:")+64 {
		t.Errorf("auth should include 64-char signature, got length %d", len(resp.Auth))
	}
}

func TestVerifyWebhook_Method(t *testing.T) {
	client := NewClient("app-123", "key-456", "secret-789", "eu")

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"test"}`)

	// Sign with client's secret
	signature := SignWebhookPayload(client.secret, timestamp, body)

	headers := http.Header{}
	headers.Set("X-Realtime-Signature", signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	valid := client.VerifyWebhook(headers, body, 0)
	if !valid {
		t.Error("expected webhook to be valid")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("app-123", "key-456", "secret-789", "eu", withBaseURL(server.URL))

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	params := TriggerParams{
		Name:    "test-event",
		Channel: "test-channel",
		Data:    `{"message":"hello"}`,
	}

	err := client.Trigger(ctx, params)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got %v", err)
	}
}
