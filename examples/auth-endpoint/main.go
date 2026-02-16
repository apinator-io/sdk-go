package main

import (
	"encoding/json"
	"log"
	"net/http"

	apinator "github.com/apinator-io/sdk-go"
)

func main() {
	client := apinator.NewClient(
		"your-app-id",
		"your-key",
		"your-secret",
		"eu",
	)

	http.HandleFunc("/realtime/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		socketID := r.FormValue("socket_id")
		channelName := r.FormValue("channel_name")

		if socketID == "" || channelName == "" {
			http.Error(w, "missing socket_id or channel_name", http.StatusBadRequest)
			return
		}

		// TODO: Authenticate the user making this request.
		// Verify that the current user is allowed to subscribe to this channel.

		var auth apinator.ChannelAuthResponse

		if channelName[:9] == "presence-" {
			// For presence channels, include user identity
			channelData := `{"user_id": "user-123", "user_info": {"name": "Alice"}}`
			auth = client.AuthenticateChannel(socketID, channelName, &channelData)
		} else {
			// For private channels
			auth = client.AuthenticateChannel(socketID, channelName, nil)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(auth)
	})

	log.Println("Auth endpoint listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
