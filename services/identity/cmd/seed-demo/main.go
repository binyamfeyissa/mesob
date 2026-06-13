// seed-demo inserts Argon2id PIN hashes for the demo accounts into mesob_identity.credentials.
// Usage: go run ./cmd/seed-demo [DB_URL]
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/identity/internal/domain"
)

var demoUsers = []string{
	"01000000-0000-7000-8000-000000000001", // SUPER_ADMIN
	"01000000-0000-7000-8000-000000000002", // ADMIN
	"01000000-0000-7000-8000-000000000003", // BRANCH_MANAGER
	"01000000-0000-7000-8000-000000000004", // AGENT
	"01000000-0000-7000-8000-000000000005", // USER (T2)
	"01000000-0000-7000-8000-000000000006", // USER (T0)
}

func main() {
	dbURL := "postgres://mesob:mesob@localhost:5433/mesob_identity?sslmode=disable"
	if len(os.Args) > 1 {
		dbURL = os.Args[1]
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	for _, userID := range demoUsers {
		hash, err := domain.HashPIN("111111")
		if err != nil {
			fmt.Fprintf(os.Stderr, "hash: %v\n", err)
			os.Exit(1)
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO credentials (user_id, pin_hash, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id) DO UPDATE SET pin_hash = EXCLUDED.pin_hash, updated_at = NOW()
		`, userID, hash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "insert %s: %v\n", userID, err)
			os.Exit(1)
		}
		fmt.Printf("seeded %s\n", userID)
	}
	fmt.Println("done — PIN 111111 for all demo accounts")
}
