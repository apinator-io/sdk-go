# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0](https://github.com/apinator-io/sdk-go/releases/tag/v1.0.0) - 2026-02-17

### Added

- `NewClient` constructor with `appID`, `key`, `secret`, `cluster` parameters
- `WithHTTPClient` functional option for custom HTTP clients
- `Client.Trigger` method for publishing events to channels
- `Client.AuthenticateChannel` method for private and presence channel auth
- `Client.GetChannels` method for listing active channels with optional prefix filter
- `Client.GetChannel` method for querying a specific channel
- `Client.VerifyWebhook` method for verifying webhook signatures
- Standalone `SignRequest` function for HMAC-SHA256 request signing
- Standalone `SignChannel` function for channel auth signatures
- Standalone `AuthenticateChannel` function for generating auth responses
- Standalone `VerifyWebhook` function for webhook signature verification
- Standalone `SignWebhookPayload` function for webhook payload signing
- `ApiError`, `AuthenticationError`, and `ValidationError` error types
- `TriggerParams`, `ChannelAuthResponse`, and `ChannelInfo` types
