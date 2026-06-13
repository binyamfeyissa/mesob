package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"time"
)

// LoggingProcessor verifies HMAC-SHA256 signatures and logs partner webhook events.
// Replace with a Kafka-emitting processor once the event catalog is finalised.
type LoggingProcessor struct{}

func (p *LoggingProcessor) Process(ctx context.Context, partner string, body io.Reader, signature, timestamp string) error {
	raw, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	// Verify timestamp skew (reject if >5 minutes old).
	if timestamp != "" {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err == nil {
			skew := math.Abs(float64(time.Now().Unix() - ts))
			if skew > 300 {
				return fmt.Errorf("webhook timestamp skew too large: %.0fs", skew)
			}
		}
	}

	// Verify HMAC-SHA256 if a secret is configured.
	secret := os.Getenv("MESOB_WEBHOOK_SECRET_" + partner)
	if secret == "" {
		secret = os.Getenv("MESOB_WEBHOOK_SECRET")
	}
	if secret != "" && signature != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(raw)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			return fmt.Errorf("webhook signature mismatch for partner %s", partner)
		}
	}

	// Log the event payload (dev/staging); in production emit to Kafka.
	var pretty map[string]any
	if err := json.Unmarshal(raw, &pretty); err == nil {
		encoded, _ := json.Marshal(pretty)
		fmt.Printf("[webhook] partner=%s payload=%s\n", partner, encoded)
	} else {
		fmt.Printf("[webhook] partner=%s raw=%s\n", partner, raw)
	}

	return nil
}
