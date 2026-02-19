package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/deletionrequest"
	"github.com/riverqueue/river"
)

// DeleteExpiredUsersArgs is the argument type for the DeleteExpiredUsers job.
type DeleteExpiredUsersArgs struct{}

func (DeleteExpiredUsersArgs) Kind() string { return "delete_expired_users" }

// DeleteExpiredUsersWorker processes pending deletion requests whose
// scheduled_at has passed, permanently deleting the user account.
type DeleteExpiredUsersWorker struct {
	river.WorkerDefaults[DeleteExpiredUsersArgs]
	db *ent.Client
}

// NewDeleteExpiredUsersWorker creates a new worker.
func NewDeleteExpiredUsersWorker(db *ent.Client) *DeleteExpiredUsersWorker {
	return &DeleteExpiredUsersWorker{db: db}
}

// Work queries all pending deletion requests where scheduled_at <= now,
// deletes the deletion request record first (to avoid FK conflicts),
// then deletes the user account (which cascades to all user data).
func (w *DeleteExpiredUsersWorker) Work(ctx context.Context, _ *river.Job[DeleteExpiredUsersArgs]) error {
	now := time.Now()

	requests, err := w.db.DeletionRequest.Query().
		Where(
			deletionrequest.StatusEQ(deletionrequest.StatusPending),
			deletionrequest.ScheduledAtLTE(now),
		).
		All(ctx)
	if err != nil {
		return err
	}

	for _, req := range requests {
		userID := req.UserID

		// Delete the deletion request first to remove the FK reference.
		// (ON DELETE CASCADE in the schema handles this automatically in
		// production, but we do it explicitly to be safe.)
		if err := w.db.DeletionRequest.DeleteOneID(req.ID).Exec(ctx); err != nil {
			slog.Error("failed to delete deletion request",
				"request_id", req.ID, "error", err)
			continue
		}

		// Delete user — Ent CASCADE removes all related data.
		if err := w.db.UserProfile.DeleteOneID(userID).Exec(ctx); err != nil {
			slog.Error("failed to delete user", "user_id", userID, "error", err)
			continue
		}

		slog.Info("user account permanently deleted (PIPA)",
			"user_id", userID, "request_id", req.ID)
	}

	return nil
}
