package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// QuotaHitEvent persists rate limit events previously tracked in memory only.
// usage_log_id is nullable to support future pre-throttle events (no call made).
type QuotaHitEvent struct {
	ent.Schema
}

func (QuotaHitEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

func (QuotaHitEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").MaxLen(20).Comment("gemini | groq"),
		field.String("model").MaxLen(100),
		field.String("feature").MaxLen(30).Optional().Nillable(),
		field.Text("error_message"),
		field.UUID("usage_log_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("FK to usage_logs, nullable for pre-throttle events"),
	}
}

func (QuotaHitEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("usage_log", UsageLog.Type).
			Field("usage_log_id").
			Unique(),
	}
}

func (QuotaHitEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "model", "created_at"),
		index.Fields("created_at"),
	}
}
