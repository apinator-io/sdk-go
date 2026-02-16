package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	apinator "github.com/apinator-io/sdk-go"
)

func main() {
	client := apinator.NewClient(
		"your-app-id",
		"your-key",
		"your-secret",
		"eu",
	)

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

		// Verify the webhook signature with a 5-minute tolerance
		if !client.VerifyWebhook(r.Header, body, 5*time.Minute) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// Parse the webhook payload
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Received webhook: %v", payload)

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Webhook handler listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
