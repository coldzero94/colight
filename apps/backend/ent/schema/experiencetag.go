package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ExperienceTag struct {
	ent.Schema
}

func (ExperienceTag) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (ExperienceTag) Fields() []ent.Field {
	return []ent.Field{
		field.String("tag_name").
			NotEmpty().
			MaxLen(50).
			Comment("Tag name (competency name)"),
		field.String("tag_type").
			NotEmpty().
			MaxLen(20).
			Comment("Tag type: skill, soft_skill, industry, keyword"),
		field.Float("confidence").
			Default(0).
			Comment("AI classification confidence (0~1)"),
	}
}

func (ExperienceTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("experience", Experience.Type).
			Ref("tags").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (ExperienceTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tag_name"),
	}
}
