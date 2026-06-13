// Package sms provides an SMS dispatch client.
// The LoggerClient prints messages to stdout for development.
// Replace with an Infobip/Twilio client for production.
package sms

import (
	"context"
	"fmt"
)

type LoggerClient struct{}

func (c *LoggerClient) Send(_ context.Context, msisdn, body string) error {
	fmt.Printf("[SMS → %s] %s\n", msisdn, body)
	return nil
}
