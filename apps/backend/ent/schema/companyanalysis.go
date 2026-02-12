package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type CompanyAnalysis struct {
	ent.Schema
}

func (CompanyAnalysis) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (CompanyAnalysis) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References auth.users(id)"),
		field.String("company_name").
			NotEmpty().
			MaxLen(100).
			Comment("Company name"),
		field.Text("job_url").
			Optional().
			Comment("Job posting URL"),
		field.JSON("job_posting", map[string]interface{}{}).
			Optional().
			Comment("Parsed job posting data"),
		field.JSON("company_info", map[string]interface{}{}).
			Optional().
			Comment("DART company info"),
		field.JSON("analysis_result", map[string]interface{}{}).
			Optional().
			Comment("AI analysis result"),
		field.JSON("matching_result", map[string]interface{}{}).
			Optional().
			Comment("Experience matching result"),
		field.Int("overall_fit_score").
			Optional().
			Nillable().
			Comment("Overall fit score (0~100)"),
	}
}

func (CompanyAnalysis) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("company_analyses").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("applications", Application.Type),
	}
}

func (CompanyAnalysis) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "company_name"),
	}
}
