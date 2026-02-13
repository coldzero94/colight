package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type ExperienceUsage struct {
	ent.Schema
}

func (ExperienceUsage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (ExperienceUsage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References user_profiles(id)"),
		field.Text("question_text").
			Optional().
			Comment("Which question this experience was used for"),
		field.String("company_name").
			Optional().
			MaxLen(100).
			Comment("Which company this experience was used for"),
		field.Time("used_at").
			Default(time.Now).
			Comment("Usage timestamp"),
	}
}

func (ExperienceUsage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("experience", Experience.Type).
			Ref("usages").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("user", UserProfile.Type).
			Ref("experience_usages").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("application", Application.Type).
			Ref("experience_usages").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.From("cover_letter", CoverLetter.Type).
			Ref("experience_usages").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (ExperienceUsage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
