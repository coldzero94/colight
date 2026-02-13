package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type CoachingSession struct {
	ent.Schema
}

func (CoachingSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (CoachingSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References user_profiles(id)"),
		field.Enum("session_type").
			Values("question_analysis", "draft", "review", "enhance").
			Comment("Session type"),
		field.JSON("input_data", map[string]interface{}{}).
			Optional().
			Comment("Input data"),
		field.JSON("output_data", map[string]interface{}{}).
			Optional().
			Comment("AI response data"),
		field.String("model_used").
			Optional().
			MaxLen(50).
			Comment("AI model used"),
		field.Int("input_tokens").
			Default(0).
			Comment("Input token count"),
		field.Int("output_tokens").
			Default(0).
			Comment("Output token count"),
		field.Float("total_cost_krw").
			Default(0).
			Comment("Cost in KRW"),
		field.Int("latency_ms").
			Default(0).
			Comment("Response latency in ms"),
		field.Float("quality_score").
			Optional().
			Nillable().
			Comment("User feedback quality score"),
	}
}

func (CoachingSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("coaching_sessions").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("cover_letter", CoverLetter.Type).
			Ref("coaching_sessions").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.From("prompt_template", PromptTemplate.Type).
			Ref("coaching_sessions").
			Unique(),
		edge.To("cover_letter_versions", CoverLetterVersion.Type),
	}
}

func (CoachingSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("session_type"),
	}
}
