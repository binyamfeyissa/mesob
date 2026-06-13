# notification

Multi-channel notification service. Handles internal sends (OTP) and Kafka event-driven dispatch.

**Port**: 8012  
**DB**: mesob_notification  
**Channels**: SMS, USSD, Voice, Telegram, Push (FCM)  
**Languages**: am (Amharic), om (Oromo), ti (Tigrinya), en (English)

## Local run
`MESOB_NOTIFICATION_DB_URL=... go run ./cmd/server`
