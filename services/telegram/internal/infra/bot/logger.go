package bot

import (
	"context"
	"fmt"
)

// LoggerClient logs bot messages to stdout. Replace with a real Telegram Bot API client.
type LoggerClient struct{}

func (c *LoggerClient) SendMessage(_ context.Context, chatID int64, text string) error {
	fmt.Printf("[TelegramBot → chat %d] %s\n", chatID, text)
	return nil
}
