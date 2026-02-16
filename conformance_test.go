package apinator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type hmacFixture struct {
	Cases []struct {
		Name              string `json:"name"`
		Secret            string `json:"secret"`
		Method            string `json:"method"`
		Path              string `json:"path"`
		Body              string `json:"body"`
		Timestamp         int64  `json:"timestamp"`
		ExpectedSignature string `json:"expected_signature"`
	} `json:"cases"`
}

type channelFixture struct {
	Cases []struct {
		Name              string  `json:"name"`
		Secret            string  `json:"secret"`
		Key               string  `json:"key"`
		SocketID          string  `json:"socket_id"`
		ChannelName       string  `json:"channel_name"`
		ChannelData       *string `json:"channel_data"`
		ExpectedSignature string  `json:"expected_signature"`
		ExpectedAuth      string  `json:"expected_auth"`
	} `json:"cases"`
}

type webhookFixture struct {
	Cases []struct {
		Name              string `json:"name"`
		Secret            string `json:"secret"`
		Timestamp         string `json:"timestamp"`
		Body              string `json:"body"`
		ExpectedSignature string `json:"expected_signature"`
	} `json:"cases"`
}

func loadFixture[T any](t *testing.T, name string) T {
	t.Helper()

	var v T
	p := filepath.Join("..", "..", "backend", "specs", "conformance", name)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("Conformance fixture not available: %s (only runs inside monorepo)", name)
		}
		t.Fatalf("reading fixture %s: %v", p, err)
	}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("parsing fixture %s: %v", p, err)
	}
	return v
}

func TestConformance_HMACVectors(t *testing.T) {
	fx := loadFixture[hmacFixture](t, "hmac-request.v1.json")
	for _, c := range fx.Cases {
		got := SignRequest(c.Secret, c.Method, c.Path, []byte(c.Body), c.Timestamp)
		if got != c.ExpectedSignature {
			t.Fatalf("case %s: expected %s, got %s", c.Name, c.ExpectedSignature, got)
		}
	}
}

func TestConformance_ChannelVectors(t *testing.T) {
	fx := loadFixture[channelFixture](t, "channel-auth.v1.json")
	for _, c := range fx.Cases {
		sig := SignChannel(c.Secret, c.SocketID, c.ChannelName, c.ChannelData)
		if sig != c.ExpectedSignature {
			t.Fatalf("case %s: expected signature %s, got %s", c.Name, c.ExpectedSignature, sig)
		}

		auth := AuthenticateChannel(c.Secret, c.Key, c.SocketID, c.ChannelName, c.ChannelData)
		if auth.Auth != c.ExpectedAuth {
			t.Fatalf("case %s: expected auth %s, got %s", c.Name, c.ExpectedAuth, auth.Auth)
		}
	}
}

func TestConformance_WebhookVectors(t *testing.T) {
	fx := loadFixture[webhookFixture](t, "webhook-signature.v1.json")
	for _, c := range fx.Cases {
		got := SignWebhookPayload(c.Secret, c.Timestamp, []byte(c.Body))
		if got != c.ExpectedSignature {
			t.Fatalf("case %s: expected signature %s, got %s", c.Name, c.ExpectedSignature, got)
		}
	}
}

func TestConformance_ParsesRFC7807(t *testing.T) {
	client := NewClient("app-123", "key-456", "secret-789", "eu")
	err := client.handleError(401, []byte(`{
		"type":"https://docs.apinator.io/problems/unauthorized",
		"title":"Unauthorized",
		"status":401,
		"detail":"signature mismatch",
		"code":"unauthorized"
	}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("expected AuthenticationError, got %T", err)
	}
	if !strings.Contains(err.Error(), "signature mismatch") {
		t.Fatalf("expected error to include problem detail, got %v", err)
	}
}

func TestConformance_ParsesRFC7807Validation(t *testing.T) {
	client := NewClient("app-123", "key-456", "secret-789", "eu")
	err := client.handleError(422, []byte(`{
		"type":"https://docs.apinator.io/problems/unprocessable_entity",
		"title":"Unprocessable Entity",
		"status":422,
		"detail":"invalid request",
		"code":"unprocessable_entity"
	}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if !strings.Contains(err.Error(), "invalid request") {
		t.Fatalf("expected error to include problem detail, got %v", err)
	}
}

func TestConformance_ParsesRFC7807GenericAPIError(t *testing.T) {
	client := NewClient("app-123", "key-456", "secret-789", "eu")
	err := client.handleError(500, []byte(`{
		"type":"https://docs.apinator.io/problems/internal_error",
		"title":"Internal Server Error",
		"status":500,
		"detail":"",
		"code":"internal_error"
	}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*ApiError); !ok {
		t.Fatalf("expected ApiError, got %T", err)
	}
	if !strings.Contains(err.Error(), "Internal Server Error") {
		t.Fatalf("expected error to include RFC7807 title fallback, got %v", err)
	}
}

func TestConformance_CanonicalPathRuleForQueryCase(t *testing.T) {
	fx := loadFixture[hmacFixture](t, "hmac-request.v1.json")
	var found bool
	for _, c := range fx.Cases {
		if c.Name != "query-not-signed" {
			continue
		}
		found = true

		canonical := SignRequest(c.Secret, c.Method, c.Path, []byte(c.Body), c.Timestamp)
		if canonical != c.ExpectedSignature {
			t.Fatalf("expected canonical path signature %s, got %s", c.ExpectedSignature, canonical)
		}

		legacy := SignRequest(c.Secret, c.Method, c.Path+"?filter_by_prefix=private-", []byte(c.Body), c.Timestamp)
		if legacy == c.ExpectedSignature {
			t.Fatal("legacy query-included signature unexpectedly matched canonical signature")
		}
	}
	if !found {
		t.Fatal("query-not-signed case missing in fixture")
	}
}
