// Package fcm provides Firebase Cloud Messaging push dispatch.
// The LoggerClient prints notifications to stdout for development.
// Replace with firebase.google.com/go/messaging in production.
package fcm

import (
	"context"
	"fmt"
)

type LoggerClient struct{}

func (c *LoggerClient) Push(_ context.Context, deviceToken, title, body string) error {
	fmt.Printf("[FCM → %s] %s: %s\n", deviceToken, title, body)
	return nil
}
