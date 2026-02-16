package apinator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// VerifyWebhook verifies a webhook signature.
// Checks the X-Realtime-Signature header (strips "sha256=" prefix if present).
// If maxAge > 0, verifies timestamp freshness.
// input = timestamp + "." + string(body)
// expected = hex(hmac-sha256(secret, input))
// Returns true if signatures match using constant-time comparison.
func VerifyWebhook(secret string, headers http.Header, body []byte, maxAge time.Duration) bool {
	// Get signature header
	sigHeader := headers.Get("X-Realtime-Signature")
	if sigHeader == "" {
		return false
	}

	// Strip "sha256=" prefix if present
	signature := strings.TrimPrefix(sigHeader, "sha256=")

	// Get timestamp header
	timestampHeader := headers.Get("X-Realtime-Timestamp")
	if timestampHeader == "" {
		return false
	}

	// Check timestamp freshness if maxAge is specified
	if maxAge > 0 {
		timestamp, err := strconv.ParseInt(timestampHeader, 10, 64)
		if err != nil {
			return false
		}

		now := time.Now().Unix()
		age := now - timestamp
		if age < 0 || age > int64(maxAge.Seconds()) {
			return false
		}
	}

	// Compute expected signature
	expected := SignWebhookPayload(secret, timestampHeader, body)

	// Decode the signature from hex for constant-time comparison
	actualBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}

	return hmac.Equal(expectedBytes, actualBytes)
}

// SignWebhookPayload signs a webhook payload.
// input = timestamp + "." + string(payload)
// Returns hex(hmac-sha256(secret, input)).
func SignWebhookPayload(secret, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
