package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/notification/internal/domain"
)

type SendUseCase struct {
	Templates  TemplateRepository
	Deliveries DeliveryRepository
	SMS        SMSClient
	FCM        FCMClient
	Telegram   TelegramClient
	Users      UserResolver
}

type SendInput struct {
	UserID      string
	TemplateKey string
	Params      map[string]string
	ChannelHint domain.Channel
}

type SendOutput struct {
	DeliveryID string
	Status     domain.DeliveryStatus
}

func (uc *SendUseCase) Execute(ctx context.Context, in SendInput) (SendOutput, error) {
	// 1. Resolve user contact info
	var msisdn string
	var lang domain.Lang = domain.LangAmharic
	var telegramChatID int64
	var fcmToken string
	if uc.Users != nil {
		var err error
		msisdn, lang, telegramChatID, fcmToken, err = uc.Users.GetContactInfo(ctx, in.UserID)
		if err != nil {
			lang = domain.LangAmharic
		}
	}

	// 2. Select channel based on hint and available contacts
	ch := in.ChannelHint
	if ch == "" {
		if msisdn != "" {
			ch = domain.ChannelSMS
		} else if telegramChatID != 0 {
			ch = domain.ChannelTelegram
		} else if fcmToken != "" {
			ch = domain.ChannelPush
		} else {
			ch = domain.ChannelSMS
		}
	}

	// 3. Find template (try exact lang, fall back to English)
	body := in.TemplateKey // fallback if no template found
	if uc.Templates != nil {
		t, err := uc.Templates.FindByKeyLangChannel(ctx, in.TemplateKey, lang, ch)
		if err != nil && lang != domain.LangEnglish {
			t, err = uc.Templates.FindByKeyLangChannel(ctx, in.TemplateKey, domain.LangEnglish, ch)
		}
		if err == nil && t != nil {
			body = t.Body
		}
	}

	// 4. Render template
	body = renderTemplate(body, in.Params)

	// 5. Dispatch to channel
	var dispatchErr error
	switch ch {
	case domain.ChannelSMS:
		if uc.SMS != nil && msisdn != "" {
			dispatchErr = uc.SMS.Send(ctx, msisdn, body)
		}
	case domain.ChannelTelegram:
		if uc.Telegram != nil && telegramChatID != 0 {
			dispatchErr = uc.Telegram.Send(ctx, telegramChatID, body)
		}
	case domain.ChannelPush:
		if uc.FCM != nil && fcmToken != "" {
			title := in.TemplateKey
			dispatchErr = uc.FCM.Push(ctx, fcmToken, title, body)
		}
	}

	status := domain.DeliveryStatusSent
	lastErr := ""
	if dispatchErr != nil {
		status = domain.DeliveryStatusFailed
		lastErr = dispatchErr.Error()
	}

	// 6. Save delivery log
	deliveryID, _ := uuid.NewV7()
	d := &domain.Delivery{
		ID:          deliveryID.String(),
		UserID:      in.UserID,
		TemplateKey: in.TemplateKey,
		Channel:     ch,
		Status:      status,
		Attempts:    1,
		LastError:   lastErr,
		CreatedAt:   time.Now().UTC(),
	}
	if uc.Deliveries != nil {
		uc.Deliveries.Save(ctx, d)
	}

	return SendOutput{DeliveryID: deliveryID.String(), Status: status}, nil
}

func renderTemplate(body string, params map[string]string) string {
	for k, v := range params {
		body = strings.ReplaceAll(body, "{{"+k+"}}", v)
	}
	return body
}

type UpsertTemplateInput struct {
	Key     string
	Lang    domain.Lang
	Channel domain.Channel
	Body    string
}

func (uc *SendUseCase) UpsertTemplate(ctx context.Context, in UpsertTemplateInput) error {
	_ = fmt.Sprintf("upserting template %s", in.Key)
	if uc.Templates == nil {
		return nil
	}
	return uc.Templates.Upsert(ctx, &domain.Template{
		Key:     in.Key,
		Lang:    in.Lang,
		Channel: in.Channel,
		Body:    in.Body,
	})
}
