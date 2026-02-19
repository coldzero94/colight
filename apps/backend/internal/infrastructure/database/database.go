package database

import (
	"context"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/lib/pq"
)

// NewClient creates an Ent client using lib/pq (database/sql).
func NewClient(databaseURL string) (*ent.Client, error) {
	return ent.Open("postgres", databaseURL)
}

// NewPool creates a pgx connection pool for River job queue.
// This is separate from the Ent client so existing code is unaffected.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, databaseURL)
}
