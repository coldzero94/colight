package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		// --- Auth fields ---
		field.String("email").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Email address"),
		field.String("password_hash").
			Optional().
			Nillable().
			MaxLen(255).
			Sensitive().
			Comment("bcrypt hashed password (email auth only)"),
		field.String("naver_id").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Naver OAuth user ID"),
		field.Enum("auth_provider").
			Values("email", "naver").
			Default("email").
			Comment("Authentication provider"),
		field.Enum("role").
			Values("user", "admin").
			Default("user").
			Comment("User role"),
		field.Bool("email_verified").
			Default(false).
			Comment("Whether email is verified"),
		field.Time("last_login_at").
			Optional().
			Nillable().
			Comment("Last login timestamp"),

		// --- Profile fields ---
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
		index.Fields("email").Unique(),
		index.Fields("naver_id").Unique(),
		index.Fields("role"),
	}
}
