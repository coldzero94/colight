package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type CoverLetter struct {
	ent.Schema
}

func (CoverLetter) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (CoverLetter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References user_profiles(id)"),
		field.String("company_name").
			Optional().
			MaxLen(100).
			Comment("Company name"),
		field.Text("question_text").
			NotEmpty().
			Comment("Cover letter question"),
		field.JSON("question_analysis", map[string]interface{}{}).
			Optional().
			Comment("Question analysis result"),
		field.Int("char_limit").
			Optional().
			Nillable().
			Comment("Character limit"),
		field.Text("current_content").
			Optional().
			Comment("Current content"),
		field.Enum("status").
			Values("draft", "reviewing", "completed").
			Default("draft").
			Comment("Cover letter status"),
	}
}

func (CoverLetter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("cover_letters").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("application", Application.Type).
			Ref("cover_letters").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("versions", CoverLetterVersion.Type),
		edge.To("coaching_sessions", CoachingSession.Type),
		edge.To("experience_usages", ExperienceUsage.Type),
	}
}

func (CoverLetter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "status"),
	}
}
