package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CompanyAnalysisCache struct {
	ent.Schema
}

func (CompanyAnalysisCache) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (CompanyAnalysisCache) Fields() []ent.Field {
	return []ent.Field{
		field.String("cache_key").
			Unique().
			NotEmpty().
			MaxLen(255).
			Comment("URL hash or company name based key"),
		field.String("cache_type").
			NotEmpty().
			MaxLen(30).
			Comment("Cache type: job_posting, company_info, full_analysis"),
		field.JSON("data", map[string]interface{}{}).
			Comment("Cached data"),
		field.Text("source_url").
			Optional().
			Comment("Original URL"),
		field.String("company_name").
			Optional().
			MaxLen(100).
			Comment("Company name"),
		field.Time("expires_at").
			Comment("Expiration time (created_at + 7 days)"),
	}
}

func (CompanyAnalysisCache) Edges() []ent.Edge {
	return nil
}

func (CompanyAnalysisCache) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cache_key"),
		index.Fields("expires_at"),
		index.Fields("company_name"),
	}
}
