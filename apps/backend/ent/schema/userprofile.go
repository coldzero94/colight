package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type UserProfile struct {
	ent.Schema
}

func (UserProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (UserProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Unique().
			Comment("References auth.users(id)"),
		field.String("nickname").
			Optional().
			MaxLen(50).
			Comment("Display name"),
		field.String("target_job").
			Optional().
			MaxLen(100).
			Comment("Target job position"),
		field.String("target_industry").
			Optional().
			MaxLen(100).
			Comment("Target industry"),
		field.String("education_level").
			Optional().
			MaxLen(20).
			Comment("Education level: 고졸, 대졸, 석사, 박사"),
		field.Int("graduation_year").
			Optional().
			Nillable().
			Comment("Graduation year"),
		field.Int("experience_years").
			Default(0).
			Comment("Years of experience (0 for new grad)"),
		field.Bool("onboarding_completed").
			Default(false).
			Comment("Whether onboarding is completed"),
	}
}

func (UserProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("experiences", Experience.Type),
		edge.To("applications", Application.Type),
		edge.To("company_analyses", CompanyAnalysis.Type),
		edge.To("cover_letters", CoverLetter.Type),
		edge.To("coaching_sessions", CoachingSession.Type),
		edge.To("experience_usages", ExperienceUsage.Type),
	}
}

func (UserProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
