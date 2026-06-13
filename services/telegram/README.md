# telegram

Telegram Bot service for Mesob Wallet. Receives Bot API webhooks and forwards commands to the gateway.

**Port**: 8011  
**Auth**: Telegram webhook secret validation

## Local run
`MESOB_TELEGRAM_BOT_TOKEN=xxx go run ./cmd/server`
