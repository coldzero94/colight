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
		field.String("provider").
			MaxLen(20).
			Optional().
			Nillable().
			Comment("AI provider: anthropic, gemini, groq"),
		field.String("model").
			MaxLen(100).
			Optional().
			Nillable().
			Comment("AI model name used"),
		field.Int("input_tokens").
			Default(0).
			Comment("Input token count"),
		field.Int("output_tokens").
			Default(0).
			Comment("Output token count"),
		field.Int("total_tokens").
			Default(0).
			Comment("Total token count"),
		field.Float("estimated_cost_krw").
			Optional().
			Nillable().
			Comment("Estimated cost in KRW"),
		field.Int("latency_ms").
			Optional().
			Nillable().
			Comment("API call latency in milliseconds"),
		field.String("status").
			MaxLen(20).
			Default("success").
			Comment("Call status: success, error"),
		field.Text("error_message").
			Optional().
			Nillable().
			Comment("Error message if status=error"),
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
