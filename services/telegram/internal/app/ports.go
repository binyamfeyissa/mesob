package app

import "context"

type BotClient interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type GatewayClient interface {
	ForwardCommand(ctx context.Context, msisdn string, command string, payload map[string]string) (string, error)
}
