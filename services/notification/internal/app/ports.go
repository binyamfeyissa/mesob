package app

import (
	"context"

	"github.com/mesob-wallet/notification/internal/domain"
)

type TemplateRepository interface {
	FindByKeyLangChannel(ctx context.Context, key string, lang domain.Lang, channel domain.Channel) (*domain.Template, error)
	Upsert(ctx context.Context, t *domain.Template) error
}

type DeliveryRepository interface {
	Save(ctx context.Context, d *domain.Delivery) error
	FindByID(ctx context.Context, id string) (*domain.Delivery, error)
	UpdateStatus(ctx context.Context, id string, status domain.DeliveryStatus, lastErr string) error
}

type SMSClient interface {
	Send(ctx context.Context, msisdn, body string) error
}

type FCMClient interface {
	Push(ctx context.Context, deviceToken, title, body string) error
}

type TelegramClient interface {
	Send(ctx context.Context, chatID int64, text string) error
}

type UserResolver interface {
	GetContactInfo(ctx context.Context, userID string) (msisdn string, lang domain.Lang, telegramChatID int64, fcmToken string, err error)
}
