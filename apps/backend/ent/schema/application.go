package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Application struct {
	ent.Schema
}

func (Application) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Application) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References user_profiles(id)"),
		field.String("company_name").
			NotEmpty().
			MaxLen(100).
			Comment("Company name"),
		field.String("position").
			Optional().
			MaxLen(200).
			Comment("Position title"),
		field.Text("job_url").
			Optional().
			Comment("Job posting URL"),
		field.Enum("status").
			Values("preparing", "submitted", "in_review", "interview", "accepted", "rejected").
			Default("preparing").
			Comment("Application status"),
		field.Time("deadline").
			Optional().
			Nillable().
			Comment("Application deadline"),
		field.Time("applied_at").
			Optional().
			Nillable().
			Comment("Submission date"),
		field.Text("notes").
			Optional().
			Comment("Notes"),
		field.JSON("tags", []string{}).
			Optional().
			Comment("User-defined tags for filtering"),
	}
}

func (Application) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("applications").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("analysis", CompanyAnalysis.Type).
			Ref("applications").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("cover_letters", CoverLetter.Type),
		edge.To("experience_usages", ExperienceUsage.Type),
	}
}

func (Application) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "status"),
		index.Fields("user_id", "deadline"),
		index.Fields("user_id", "company_name"),
	}
}
