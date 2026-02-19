// Package scheduler manages background job processing via River.
// River runs embedded in the API server process using pgx/v5.
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/jobs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// New initialises River: applies DB migrations, registers workers,
// and configures periodic jobs. Returns a started River client.
//
// Caller must defer client.Stop(ctx) to gracefully drain workers on shutdown.
func New(ctx context.Context, pool *pgxpool.Pool, db *ent.Client) (*river.Client[pgx.Tx], error) {
	// Apply River schema migrations on startup (idempotent).
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return nil, err
	}
	res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return nil, err
	}
	for _, m := range res.Versions {
		slog.Info("river migration applied", "version", m.Version)
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, jobs.NewDeleteExpiredUsersWorker(db))

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 5},
		},
		Workers: workers,
		PeriodicJobs: []*river.PeriodicJob{
			// Run every 24 hours. RunOnStart=true ensures the first run
			// happens on server startup (catches up any delayed deletions).
			river.NewPeriodicJob(
				river.PeriodicInterval(24*time.Hour),
				func() (river.JobArgs, *river.InsertOpts) {
					return jobs.DeleteExpiredUsersArgs{}, nil
				},
				&river.PeriodicJobOpts{RunOnStart: true},
			),
		},
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
