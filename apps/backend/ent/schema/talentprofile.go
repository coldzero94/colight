package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TalentProfile struct {
	ent.Schema
}

func (TalentProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (TalentProfile) Fields() []ent.Field {
	return []ent.Field{
		field.String("company_name").
			NotEmpty().
			MaxLen(100).
			Comment("Company name"),
		field.String("industry").
			Optional().
			MaxLen(50).
			Comment("Industry"),
		field.JSON("core_values", []map[string]string{}).
			Optional().
			Comment("Core values: [{keyword, description}]"),
		field.JSON("talent_traits", []map[string]string{}).
			Optional().
			Comment("Talent traits: [{trait, description}]"),
		field.JSON("culture_keywords", []string{}).
			Optional().
			Comment("Culture keywords"),
		field.Text("source").
			Optional().
			Comment("Data source"),
		field.Bool("verified").
			Default(false).
			Comment("Whether data is verified"),
	}
}

func (TalentProfile) Edges() []ent.Edge {
	return nil
}

func (TalentProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("company_name"),
	}
}
