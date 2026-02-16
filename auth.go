package apinator

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// SignRequest creates an HMAC-SHA256 signature for an API request.
// If body is empty (len == 0), bodyMD5 is "" (NOT md5 of empty bytes).
// Otherwise bodyMD5 = hex(md5(body)).
// sigString = "{timestamp}\n{method}\n{path}\n{bodyMD5}"
// Returns hex(hmac-sha256(secret, sigString)).
func SignRequest(secret, method, path string, body []byte, timestamp int64) string {
	var bodyMD5 string
	if len(body) > 0 {
		hash := md5.Sum(body)
		bodyMD5 = hex.EncodeToString(hash[:])
	}

	sigString := fmt.Sprintf("%d\n%s\n%s\n%s", timestamp, method, path, bodyMD5)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigString))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignChannel creates an auth signature for channel subscriptions.
// sigString = socketID + ":" + channelName
// If channelData != nil: sigString += ":" + *channelData
// Returns hex(hmac-sha256(secret, sigString)).
func SignChannel(secret, socketID, channelName string, channelData *string) string {
	sigString := socketID + ":" + channelName
	if channelData != nil {
		sigString += ":" + *channelData
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigString))
	return hex.EncodeToString(mac.Sum(nil))
}

// AuthenticateChannel returns the auth response for a channel subscription.
// The auth field is formatted as "{key}:{signature}".
func AuthenticateChannel(secret, key, socketID, channelName string, channelData *string) ChannelAuthResponse {
	sig := SignChannel(secret, socketID, channelName, channelData)
	resp := ChannelAuthResponse{
		Auth: key + ":" + sig,
	}
	if channelData != nil {
		resp.ChannelData = channelData
	}
	return resp
}
