package apinator

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestVerifyWebhook_Valid(t *testing.T) {
	secret := "my-secret-key"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Sign the payload
	signature := SignWebhookPayload(secret, timestamp, body)

	// Create headers
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "sha256="+signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify
	valid := VerifyWebhook(secret, headers, body, 0)
	if !valid {
		t.Error("expected webhook to be valid")
	}
}

func TestVerifyWebhook_ValidWithoutPrefix(t *testing.T) {
	secret := "my-secret-key"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Sign the payload
	signature := SignWebhookPayload(secret, timestamp, body)

	// Create headers without sha256= prefix
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify
	valid := VerifyWebhook(secret, headers, body, 0)
	if !valid {
		t.Error("expected webhook to be valid")
	}
}

func TestVerifyWebhook_TamperedBody(t *testing.T) {
	secret := "my-secret-key"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	originalBody := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)
	tamperedBody := []byte(`{"event":"channel_occupied","channel":"hacked-channel"}`)

	// Sign the original payload
	signature := SignWebhookPayload(secret, timestamp, originalBody)

	// Create headers
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "sha256="+signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify with tampered body
	valid := VerifyWebhook(secret, headers, tamperedBody, 0)
	if valid {
		t.Error("expected webhook to be invalid with tampered body")
	}
}

func TestVerifyWebhook_WrongSecret(t *testing.T) {
	secret := "my-secret-key"
	wrongSecret := "wrong-secret"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Sign with wrong secret
	signature := SignWebhookPayload(wrongSecret, timestamp, body)

	// Create headers
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "sha256="+signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify with correct secret
	valid := VerifyWebhook(secret, headers, body, 0)
	if valid {
		t.Error("expected webhook to be invalid with wrong secret")
	}
}

func TestVerifyWebhook_ExpiredTimestamp(t *testing.T) {
	secret := "my-secret-key"
	// Timestamp from 10 minutes ago
	pastTime := time.Now().Add(-10 * time.Minute)
	timestamp := strconv.FormatInt(pastTime.Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Sign the payload
	signature := SignWebhookPayload(secret, timestamp, body)

	// Create headers
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "sha256="+signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify with 5-minute max age
	maxAge := 5 * time.Minute
	valid := VerifyWebhook(secret, headers, body, maxAge)
	if valid {
		t.Error("expected webhook to be invalid due to expired timestamp")
	}
}

func TestVerifyWebhook_FreshTimestamp(t *testing.T) {
	secret := "my-secret-key"
	// Current timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Sign the payload
	signature := SignWebhookPayload(secret, timestamp, body)

	// Create headers
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "sha256="+signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify with 5-minute max age
	maxAge := 5 * time.Minute
	valid := VerifyWebhook(secret, headers, body, maxAge)
	if !valid {
		t.Error("expected webhook to be valid with fresh timestamp")
	}
}

func TestVerifyWebhook_FutureTimestamp(t *testing.T) {
	secret := "my-secret-key"
	// Timestamp from future (should be invalid)
	futureTime := time.Now().Add(10 * time.Minute)
	timestamp := strconv.FormatInt(futureTime.Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Sign the payload
	signature := SignWebhookPayload(secret, timestamp, body)

	// Create headers
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "sha256="+signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify with 5-minute max age (future timestamps should be rejected)
	maxAge := 5 * time.Minute
	valid := VerifyWebhook(secret, headers, body, maxAge)
	if valid {
		t.Error("expected webhook to be invalid with future timestamp")
	}
}

func TestVerifyWebhook_MissingSignature(t *testing.T) {
	secret := "my-secret-key"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Create headers without signature
	headers := http.Header{}
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify
	valid := VerifyWebhook(secret, headers, body, 0)
	if valid {
		t.Error("expected webhook to be invalid without signature")
	}
}

func TestVerifyWebhook_MissingTimestamp(t *testing.T) {
	secret := "my-secret-key"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Sign the payload
	signature := SignWebhookPayload(secret, timestamp, body)

	// Create headers without timestamp
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "sha256="+signature)

	// Verify
	valid := VerifyWebhook(secret, headers, body, 0)
	if valid {
		t.Error("expected webhook to be invalid without timestamp")
	}
}

func TestVerifyWebhook_InvalidSignatureFormat(t *testing.T) {
	secret := "my-secret-key"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"event":"channel_occupied","channel":"test-channel"}`)

	// Create headers with invalid signature (not hex)
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", "not-a-valid-hex-string!!!")
	headers.Set("X-Realtime-Timestamp", timestamp)

	// Verify
	valid := VerifyWebhook(secret, headers, body, 0)
	if valid {
		t.Error("expected webhook to be invalid with malformed signature")
	}
}

func TestSignWebhookPayload(t *testing.T) {
	secret := "my-secret-key"
	timestamp := "1700000000"
	payload := []byte(`{"event":"test"}`)

	signature := SignWebhookPayload(secret, timestamp, payload)

	// Verify signature format (hex-encoded SHA256 = 64 chars)
	if len(signature) != 64 {
		t.Errorf("expected signature length 64, got %d", len(signature))
	}

	// Verify consistency
	signature2 := SignWebhookPayload(secret, timestamp, payload)
	if signature != signature2 {
		t.Error("same inputs should produce same signature")
	}

	// Verify different inputs produce different signatures
	signature3 := SignWebhookPayload("different-secret", timestamp, payload)
	if signature == signature3 {
		t.Error("different secret should produce different signature")
	}

	signature4 := SignWebhookPayload(secret, "1700000001", payload)
	if signature == signature4 {
		t.Error("different timestamp should produce different signature")
	}

	signature5 := SignWebhookPayload(secret, timestamp, []byte(`{"event":"other"}`))
	if signature == signature5 {
		t.Error("different payload should produce different signature")
	}
}

func TestWebhook_CrossVerification(t *testing.T) {
	// These exact values should be used across all SDK implementations
	// for cross-verification
	secret := "my-secret-key"
	timestamp := "1700000000"
	payload := []byte(`{"event":"channel_occupied","channel":"test-channel","data":{}}`)

	signature := SignWebhookPayload(secret, timestamp, payload)

	// Log the signature for cross-SDK verification
	t.Logf("Cross-verification webhook signature: %s", signature)
	t.Logf("Cross-verification timestamp: %s", timestamp)
	t.Logf("Cross-verification payload: %s", string(payload))

	// Verify it
	headers := http.Header{}
	headers.Set("X-Realtime-Signature", signature)
	headers.Set("X-Realtime-Timestamp", timestamp)

	valid := VerifyWebhook(secret, headers, payload, 0)
	if !valid {
		t.Error("cross-verification webhook should be valid")
	}
}
