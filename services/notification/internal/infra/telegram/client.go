// Package telegram provides Telegram Bot API message dispatch.
// The LoggerClient prints messages to stdout for development.
// Replace with a real bot token + HTTP call in production.
package telegram

import (
	"context"
	"fmt"
)

type LoggerClient struct{}

func (c *LoggerClient) Send(_ context.Context, chatID int64, text string) error {
	fmt.Printf("[Telegram → chat=%d] %s\n", chatID, text)
	return nil
}
