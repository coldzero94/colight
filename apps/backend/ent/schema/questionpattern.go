package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type QuestionPattern struct {
	ent.Schema
}

func (QuestionPattern) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (QuestionPattern) Fields() []ent.Field {
	return []ent.Field{
		field.String("pattern_type").
			NotEmpty().
			MaxLen(30).
			Comment("Pattern type: growth, crisis, leadership, etc."),
		field.String("pattern_name").
			NotEmpty().
			MaxLen(100).
			Comment("Pattern display name"),
		field.JSON("detection_keywords", []string{}).
			Optional().
			Comment("Keywords for question detection"),
		field.Text("detection_regex").
			Optional().
			Comment("Regex pattern for detection"),
		field.JSON("primary_weapons", []string{}).
			Optional().
			Comment("Primary weapon codes array"),
		field.JSON("secondary_weapons", []string{}).
			Optional().
			Comment("Secondary weapon codes array"),
		field.JSON("writing_guide", map[string]interface{}{}).
			Optional().
			Comment("Writing guide: structure, ratios, tips"),
		field.Int("display_order").
			Default(0).
			Comment("Display order"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether this pattern is active"),
	}
}

func (QuestionPattern) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("coaching_prompt", PromptTemplate.Type).
			Ref("question_patterns").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (QuestionPattern) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("pattern_type"),
		index.Fields("is_active"),
	}
}
