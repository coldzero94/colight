package controller

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/testutil"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	// Clean DB before running this package's tests
	client, err := ent.Open("postgres", testutil.TestDSN())
	if err != nil {
		log.Fatalf("open test db for cleanup: %v", err)
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("migrate test db: %v", err)
	}
	testutil.CleanAllTables(client)
	client.Close()

	os.Exit(m.Run())
}
