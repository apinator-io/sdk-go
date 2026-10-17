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

// WebhookMaxFutureSkew is how far in the future a webhook timestamp may be
// before VerifyWebhook rejects it, when maxAge > 0. It allows for clock skew
// between Apinator and the receiver.
const WebhookMaxFutureSkew = 300 * time.Second

// VerifyWebhook verifies a webhook signature.
//
// secret is the webhook endpoint's signing secret (whsec_..., shown in the
// console per endpoint), not the app's API secret.
//
// Checks the X-Realtime-Signature header (strips "sha256=" prefix if present).
// If maxAge > 0 (5*time.Minute recommended), rejects timestamps older than
// maxAge or more than WebhookMaxFutureSkew (300s) in the future.
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
		if age > int64(maxAge.Seconds()) || -age > int64(WebhookMaxFutureSkew.Seconds()) {
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
