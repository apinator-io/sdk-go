package apinator

// TriggerParams represents the parameters for triggering an event.
type TriggerParams struct {
	Name     string   `json:"name"`
	Channel  string   `json:"channel,omitempty"`
	Channels []string `json:"channels,omitempty"`
	Data     string   `json:"data"`
	SocketID string   `json:"socket_id,omitempty"`
}

// ChannelAuthResponse represents the authentication response for a channel subscription.
type ChannelAuthResponse struct {
	Auth        string  `json:"auth"`
	ChannelData *string `json:"channel_data,omitempty"`
}

// ChannelInfo represents information about a channel.
type ChannelInfo struct {
	Name              string `json:"name"`
	SubscriptionCount int    `json:"subscription_count"`
}
