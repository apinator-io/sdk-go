package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	apinator "github.com/apinator-io/sdk-go"
)

// webhookPayload is the JSON body of every Apinator webhook delivery.
type webhookPayload struct {
	ID      string `json:"id"`    // event id: deliveries are at-least-once, dedupe on it
	Event   string `json:"event"` // channel_occupied, channel_vacated, member_added, member_removed (console test: webhook.test)
	AppID   string `json:"app_id"`
	Channel string `json:"channel"`
	UserID  string `json:"user_id"` // only on member_added / member_removed
	TimeMs  int64  `json:"time_ms"` // unix milliseconds
}

func main() {
	client := apinator.NewClient(
		os.Getenv("APINATOR_APP_ID"),
		os.Getenv("APINATOR_KEY"),
		os.Getenv("APINATOR_SECRET"),
		"eu",
	)

	// Each webhook endpoint has its own signing secret (whsec_..., shown in the
	// console). It is not the app's API secret.
	webhookSecret := os.Getenv("APINATOR_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Fatal("APINATOR_WEBHOOK_SECRET is not set")
	}

	// Event ids already handled. Use a persistent store in production.
	var mu sync.Mutex
	processed := map[string]bool{}

	http.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		// Reject bad signatures, and timestamps older than 5 minutes or more
		// than 5 minutes in the future.
		if !client.VerifyWebhook(webhookSecret, r.Header, body, 5*time.Minute) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var payload webhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		mu.Lock()
		duplicate := processed[payload.ID]
		processed[payload.ID] = true
		mu.Unlock()
		if duplicate {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch payload.Event {
		case "member_added", "member_removed":
			log.Printf("%s: user %s on %s", payload.Event, payload.UserID, payload.Channel)
		default:
			log.Printf("%s on %s", payload.Event, payload.Channel)
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Webhook handler listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
