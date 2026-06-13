package domain

import (
	"strings"
	"time"
)

type WebhookUpdate struct {
	UpdateID int64
	Message  *TelegramMessage
}

type TelegramMessage struct {
	MessageID int64
	ChatID    int64
	Text      string
	From      TelegramUser
	Date      time.Time
}

type TelegramUser struct {
	ID        int64
	Username  string
	FirstName string
}

type Command struct {
	Name    string
	Payload string
}

func ParseCommand(text string) *Command {
	if len(text) == 0 || text[0] != '/' {
		return nil
	}
	// Split "/command payload" or "/command@botname payload"
	body := text[1:]
	// Strip @botname suffix
	if at := indexOf(body, '@'); at >= 0 {
		body = body[:at]
		if sp := indexOf(text[1:], ' '); sp >= 0 {
			body = text[1:at]
			payload := strings.TrimSpace(text[at+1:])
			if sp2 := indexOf(payload, ' '); sp2 >= 0 {
				return &Command{Name: body, Payload: strings.TrimSpace(payload[sp2:])}
			}
		}
	}
	if sp := indexOf(body, ' '); sp >= 0 {
		return &Command{Name: body[:sp], Payload: strings.TrimSpace(body[sp+1:])}
	}
	return &Command{Name: body}
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
