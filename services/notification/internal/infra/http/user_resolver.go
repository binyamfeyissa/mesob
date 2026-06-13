package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mesob-wallet/notification/internal/domain"
)

// IdentityUserResolver fetches user contact info from the identity service.
// Called on the notification send path to resolve MSISDN, lang, telegram chat ID, FCM token.
type IdentityUserResolver struct {
	IdentityURL string
	HTTPClient  *http.Client
}

func NewIdentityUserResolver(identityURL string) *IdentityUserResolver {
	return &IdentityUserResolver{
		IdentityURL: identityURL,
		HTTPClient:  &http.Client{},
	}
}

type userInfoResponse struct {
	Data struct {
		UserID         string `json:"user_id"`
		MSISDN         string `json:"msisdn"`
		PreferredLang  string `json:"preferred_lang"`
		TelegramChatID int64  `json:"telegram_chat_id"`
		FCMToken       string `json:"fcm_token"`
	} `json:"data"`
}

func (r *IdentityUserResolver) GetContactInfo(
	ctx context.Context,
	userID string,
) (msisdn string, lang domain.Lang, telegramChatID int64, fcmToken string, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/identity/users/%s", r.IdentityURL, userID),
		nil,
	)
	if err != nil {
		return "", domain.LangAmharic, 0, "", err
	}

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return "", domain.LangAmharic, 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", domain.LangAmharic, 0, "", fmt.Errorf("identity returned %d for user %s", resp.StatusCode, userID)
	}

	var body userInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", domain.LangAmharic, 0, "", err
	}

	d := body.Data
	l := domain.Lang(d.PreferredLang)
	if l == "" {
		l = domain.LangAmharic
	}
	return d.MSISDN, l, d.TelegramChatID, d.FCMToken, nil
}
