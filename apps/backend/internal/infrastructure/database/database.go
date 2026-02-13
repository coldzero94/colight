package database

import (
	"github.com/coby/colight/apps/backend/ent"

	_ "github.com/lib/pq"
)

func NewClient(databaseURL string) (*ent.Client, error) {
	return ent.Open("postgres", databaseURL)
}
