package app

import (
	"context"
	"fmt"

	"github.com/mesob-wallet/telegram/internal/domain"
)

type WebhookUseCase struct {
	Bot     BotClient
	Gateway GatewayClient
}

func NewWebhookUseCase(bot BotClient, gw GatewayClient) *WebhookUseCase {
	return &WebhookUseCase{Bot: bot, Gateway: gw}
}

type WebhookInput struct {
	Update domain.WebhookUpdate
}

type WebhookOutput struct {
	Handled bool
}

func (uc *WebhookUseCase) Execute(ctx context.Context, in WebhookInput) (WebhookOutput, error) {
	msg := in.Update.Message
	if msg == nil {
		return WebhookOutput{Handled: false}, nil
	}

	cmd := domain.ParseCommand(msg.Text)
	if cmd == nil {
		// Non-command message: send help text.
		uc.reply(ctx, msg.ChatID, helpText())
		return WebhookOutput{Handled: true}, nil
	}

	switch cmd.Name {
	case "start", "help":
		uc.reply(ctx, msg.ChatID, helpText())

	case "balance", "bal":
		result := uc.forward(ctx, msg.ChatID, "BALANCE", map[string]string{})
		uc.reply(ctx, msg.ChatID, result)

	case "send":
		result := uc.forward(ctx, msg.ChatID, "SEND", map[string]string{"args": cmd.Payload})
		uc.reply(ctx, msg.ChatID, result)

	case "history":
		result := uc.forward(ctx, msg.ChatID, "HISTORY", map[string]string{})
		uc.reply(ctx, msg.ChatID, result)

	case "float":
		result := uc.forward(ctx, msg.ChatID, "FLOAT", map[string]string{})
		uc.reply(ctx, msg.ChatID, result)

	default:
		uc.reply(ctx, msg.ChatID, fmt.Sprintf("Unknown command /%s. Type /help for available commands.", cmd.Name))
	}

	return WebhookOutput{Handled: true}, nil
}

func (uc *WebhookUseCase) reply(ctx context.Context, chatID int64, text string) {
	if uc.Bot == nil {
		return
	}
	_ = uc.Bot.SendMessage(ctx, chatID, text)
}

func (uc *WebhookUseCase) forward(ctx context.Context, chatID int64, command string, payload map[string]string) string {
	if uc.Gateway == nil {
		return "Service temporarily unavailable."
	}
	// MSISDN is not known from Telegram chat ID alone — gateway must resolve it.
	result, err := uc.Gateway.ForwardCommand(ctx, fmt.Sprintf("tg:%d", chatID), command, payload)
	if err != nil {
		return "Could not process your request. Please try again."
	}
	return result
}

func helpText() string {
	return `Mesob Wallet — available commands:
/balance — check your balance
/send <msisdn> <amount> — send money
/history — recent transactions
/float — agent float status
/help — show this message`
}
