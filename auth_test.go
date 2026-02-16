package apinator

import (
	"testing"
)

func TestSignRequest_EmptyBody(t *testing.T) {
	secret := "my-secret-key"
	method := "GET"
	path := "/apps/123/channels"
	body := []byte{}
	timestamp := int64(1700000000)

	// With empty body, bodyMD5 should be ""
	signature := SignRequest(secret, method, path, body, timestamp)

	// Verify signature is not empty
	if signature == "" {
		t.Error("signature should not be empty")
	}

	// Verify signature is hex-encoded (64 chars for SHA256)
	if len(signature) != 64 {
		t.Errorf("expected signature length 64, got %d", len(signature))
	}
}

func TestSignRequest_WithBody(t *testing.T) {
	secret := "my-secret-key"
	method := "POST"
	path := "/apps/123/events"
	body := []byte(`{"name":"test"}`)
	timestamp := int64(1700000000)

	signature := SignRequest(secret, method, path, body, timestamp)

	// Verify signature format (hex-encoded SHA256 = 64 chars)
	if len(signature) != 64 {
		t.Errorf("expected signature length 64, got %d", len(signature))
	}

	// Verify consistency - same inputs produce same output
	signature2 := SignRequest(secret, method, path, body, timestamp)
	if signature != signature2 {
		t.Error("same inputs should produce same signature")
	}

	// Verify different secret produces different signature
	signature3 := SignRequest("different-secret", method, path, body, timestamp)
	if signature == signature3 {
		t.Error("different secret should produce different signature")
	}
}

func TestSignChannel_Private(t *testing.T) {
	secret := "my-secret-key"
	socketID := "12345.67890"
	channelName := "private-chat"

	signature := SignChannel(secret, socketID, channelName, nil)

	// Verify signature format
	if len(signature) != 64 {
		t.Errorf("expected signature length 64, got %d", len(signature))
	}

	// Verify consistency
	signature2 := SignChannel(secret, socketID, channelName, nil)
	if signature != signature2 {
		t.Error("same inputs should produce same signature")
	}

	// Verify different inputs produce different signatures
	signature3 := SignChannel(secret, "different-socket", channelName, nil)
	if signature == signature3 {
		t.Error("different socketID should produce different signature")
	}

	signature4 := SignChannel(secret, socketID, "private-other", nil)
	if signature == signature4 {
		t.Error("different channel should produce different signature")
	}
}

func TestSignChannel_Presence(t *testing.T) {
	secret := "my-secret-key"
	socketID := "12345.67890"
	channelName := "presence-chat"
	channelData := `{"user_id":"user1"}`

	signature := SignChannel(secret, socketID, channelName, &channelData)

	// Verify signature format
	if len(signature) != 64 {
		t.Errorf("expected signature length 64, got %d", len(signature))
	}

	// Verify consistency
	signature2 := SignChannel(secret, socketID, channelName, &channelData)
	if signature != signature2 {
		t.Error("same inputs should produce same signature")
	}

	// Verify signature with channelData differs from without
	signatureWithout := SignChannel(secret, socketID, channelName, nil)
	if signature == signatureWithout {
		t.Error("signature with channelData should differ from without")
	}

	// Verify different channelData produces different signature
	differentData := `{"user_id":"user2"}`
	signature3 := SignChannel(secret, socketID, channelName, &differentData)
	if signature == signature3 {
		t.Error("different channelData should produce different signature")
	}
}

func TestAuthenticateChannel_Private(t *testing.T) {
	secret := "my-secret-key"
	key := "app-key-123"
	socketID := "12345.67890"
	channelName := "private-chat"

	resp := AuthenticateChannel(secret, key, socketID, channelName, nil)

	// Verify auth format: {key}:{signature}
	if len(resp.Auth) < len(key)+1+64 {
		t.Errorf("auth should be at least %d chars, got %d", len(key)+1+64, len(resp.Auth))
	}

	// Verify auth starts with key
	expectedPrefix := key + ":"
	if resp.Auth[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("auth should start with %q, got %q", expectedPrefix, resp.Auth[:len(expectedPrefix)])
	}

	// Verify channelData is nil
	if resp.ChannelData != nil {
		t.Error("channelData should be nil for private channel")
	}
}

func TestAuthenticateChannel_Presence(t *testing.T) {
	secret := "my-secret-key"
	key := "app-key-123"
	socketID := "12345.67890"
	channelName := "presence-chat"
	channelData := `{"user_id":"user1"}`

	resp := AuthenticateChannel(secret, key, socketID, channelName, &channelData)

	// Verify auth format
	expectedPrefix := key + ":"
	if resp.Auth[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("auth should start with %q, got %q", expectedPrefix, resp.Auth[:len(expectedPrefix)])
	}

	// Verify channelData is included
	if resp.ChannelData == nil {
		t.Fatal("channelData should not be nil for presence channel")
	}

	if *resp.ChannelData != channelData {
		t.Errorf("expected channelData %q, got %q", channelData, *resp.ChannelData)
	}
}

func TestAuthenticateChannel_CrossVerification(t *testing.T) {
	// These exact values should be used across all SDK implementations
	// for cross-verification
	secret := "my-secret-key"
	key := "test-key"
	socketID := "12345.67890"
	channelName := "private-chat"

	resp := AuthenticateChannel(secret, key, socketID, channelName, nil)

	// Log the auth value for cross-SDK verification
	t.Logf("Cross-verification auth (private): %s", resp.Auth)

	// Test with presence
	channelDataPresence := `{"user_id":"user1"}`
	respPresence := AuthenticateChannel(secret, key, socketID, "presence-chat", &channelDataPresence)

	t.Logf("Cross-verification auth (presence): %s", respPresence.Auth)
	t.Logf("Cross-verification channelData: %s", *respPresence.ChannelData)
}
