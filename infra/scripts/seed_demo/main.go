// seed_demo inserts demo accounts into mesob_identity with correct Argon2id PIN hashes.
//
// Usage:
//   go run ./infra/scripts/seed_demo
//
// Environment variables (all optional, defaults match docker-compose):
//   MESOB_IDENTITY_DB_URL  — Postgres DSN
//
// Demo accounts created:
//   +251911000001  SUPER_ADMIN  PIN: 111111
//   +251911000002  ADMIN        PIN: 111111
//   +251911000003  BRANCH_MGR   PIN: 111111
//   +251911000004  AGENT        PIN: 111111
//   +251911000005  USER (T2)    PIN: 111111
//   +251911000006  USER (T0)    PIN: 111111

package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/argon2"
)

// Argon2id params (must match identity service credential.go)
const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

type demoAccount struct {
	id       string
	msisdn   string
	pin      string
	role     string
	kycTier  int
	regionID string
	lang     string
}

var demos = []demoAccount{
	{
		id: "01000000-0000-7000-8000-000000000001", msisdn: "+251911000001",
		pin: "111111", role: "SUPER_ADMIN", kycTier: 2,
		regionID: "00000000-0000-7000-8000-000000000001", lang: "en",
	},
	{
		id: "01000000-0000-7000-8000-000000000002", msisdn: "+251911000002",
		pin: "111111", role: "ADMIN", kycTier: 2,
		regionID: "00000000-0000-7000-8000-000000000001", lang: "en",
	},
	{
		id: "01000000-0000-7000-8000-000000000003", msisdn: "+251911000003",
		pin: "111111", role: "BRANCH_MANAGER", kycTier: 1,
		regionID: "00000000-0000-7000-8000-000000000001", lang: "am",
	},
	{
		id: "01000000-0000-7000-8000-000000000004", msisdn: "+251911000004",
		pin: "111111", role: "AGENT", kycTier: 1,
		regionID: "00000000-0000-7000-8000-000000000002", lang: "am",
	},
	{
		id: "01000000-0000-7000-8000-000000000005", msisdn: "+251911000005",
		pin: "111111", role: "USER", kycTier: 2,
		regionID: "00000000-0000-7000-8000-000000000001", lang: "am",
	},
	{
		id: "01000000-0000-7000-8000-000000000006", msisdn: "+251911000006",
		pin: "111111", role: "USER", kycTier: 0,
		regionID: "00000000-0000-7000-8000-000000000003", lang: "am",
	},
}

func hashPIN(pin string) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	hash := argon2.IDKey([]byte(pin), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	// Encode as salt_hex:hash_hex for storage
	encoded := []byte(hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash))
	return encoded, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	dsn := getenv("MESOB_IDENTITY_DB_URL", "postgres://mesob:mesob@localhost:5433/mesob_identity?sslmode=disable")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	// Ensure regions exist
	regions := []struct{ id, name, code string }{
		{"00000000-0000-7000-8000-000000000001", "Addis Ababa", "AA"},
		{"00000000-0000-7000-8000-000000000002", "Oromia", "OR"},
		{"00000000-0000-7000-8000-000000000003", "Amhara", "AM"},
		{"00000000-0000-7000-8000-000000000004", "Tigray", "TI"},
		{"00000000-0000-7000-8000-000000000005", "SNNPR", "SN"},
	}
	for _, r := range regions {
		_, err := db.ExecContext(context.Background(),
			`INSERT INTO regions (id, name, code) VALUES ($1,$2,$3) ON CONFLICT (code) DO NOTHING`,
			r.id, r.name, r.code,
		)
		if err != nil {
			log.Fatalf("insert region %s: %v", r.code, err)
		}
	}

	// Ensure role column exists
	_, err = db.ExecContext(context.Background(),
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'USER'`,
	)
	if err != nil {
		log.Fatalf("alter table: %v", err)
	}

	for _, acc := range demos {
		// Upsert user
		_, err = db.ExecContext(context.Background(), `
			INSERT INTO users (id, msisdn, kyc_tier, region_id, status, preferred_lang, role)
			VALUES ($1, $2, $3, $4, 'ACTIVE', $5, $6)
			ON CONFLICT (id) DO UPDATE SET
				kyc_tier       = EXCLUDED.kyc_tier,
				status         = EXCLUDED.status,
				preferred_lang = EXCLUDED.preferred_lang,
				role           = EXCLUDED.role,
				updated_at     = NOW()
		`, acc.id, acc.msisdn, acc.kycTier, acc.regionID, acc.lang, acc.role)
		if err != nil {
			log.Fatalf("upsert user %s: %v", acc.msisdn, err)
		}

		// Hash PIN and upsert credential
		pinHash, err := hashPIN(acc.pin)
		if err != nil {
			log.Fatalf("hash pin for %s: %v", acc.msisdn, err)
		}
		_, err = db.ExecContext(context.Background(), `
			INSERT INTO credentials (user_id, pin_hash, failed_count, updated_at)
			VALUES ($1, $2, 0, NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				pin_hash   = EXCLUDED.pin_hash,
				updated_at = NOW()
		`, acc.id, pinHash)
		if err != nil {
			log.Fatalf("upsert credential %s: %v", acc.msisdn, err)
		}

		fmt.Printf("✓ %s  %-20s  role=%-15s  kyc_tier=%d\n", acc.msisdn, acc.id, acc.role, acc.kycTier)
	}

	fmt.Println("\nDemo seed complete. All accounts use PIN: 111111")
}
