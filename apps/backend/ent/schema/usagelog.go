package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UsageLog tracks feature usage for freemium limit enforcement.
type UsageLog struct {
	ent.Schema
}

func (UsageLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (UsageLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References user_profiles(id)"),
		field.String("feature").
			MaxLen(30).
			Comment("Feature name: experience, analysis, question_analysis, draft, review"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Additional context (e.g. cover_letter_id)"),
	}
}

func (UsageLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("usage_logs").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (UsageLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "feature", "created_at"),
	}
}
