package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AICallError stores detailed error information for failed AI calls.
// It is a 1:0..1 extension of UsageLog — only populated when status=error.
type AICallError struct {
	ent.Schema
}

func (AICallError) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

func (AICallError) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("usage_log_id", uuid.UUID{}).
			Unique().
			Comment("1:1 FK to usage_logs — CASCADE on delete"),
		field.Enum("error_type").
			Values("rate_limit", "timeout", "provider_error",
				"invalid_request", "context_exceeded", "unknown").
			Comment("Classified AI error category"),
		field.Text("error_message").
			Comment("Raw error message from provider"),
	}
}

func (AICallError) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("usage_log", UsageLog.Type).
			Ref("error_detail").
			Field("usage_log_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (AICallError) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("error_type", "created_at"),
		index.Fields("created_at"),
	}
}
