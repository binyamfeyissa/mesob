package domain

import "time"

type Channel string

const (
	ChannelSMS      Channel = "SMS"
	ChannelUSSD     Channel = "USSD"
	ChannelVoice    Channel = "VOICE"
	ChannelTelegram Channel = "TELEGRAM"
	ChannelPush     Channel = "PUSH"
)

type Lang string

const (
	LangAmharic  Lang = "am"
	LangOromo    Lang = "om"
	LangTigrinya Lang = "ti"
	LangEnglish  Lang = "en"
)

type Template struct {
	Key     string
	Lang    Lang
	Channel Channel
	Body    string
}

type Delivery struct {
	ID          string
	UserID      string
	TemplateKey string
	Channel     Channel
	Status      DeliveryStatus
	Attempts    int
	LastError   string
	CreatedAt   time.Time
}

type DeliveryStatus string

const (
	DeliveryStatusQueued     DeliveryStatus = "QUEUED"
	DeliveryStatusSent       DeliveryStatus = "SENT"
	DeliveryStatusFailed     DeliveryStatus = "FAILED"
	DeliveryStatusDeadLetter DeliveryStatus = "DEAD_LETTER"
)
