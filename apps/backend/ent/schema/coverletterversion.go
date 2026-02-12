package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CoverLetterVersion struct {
	ent.Schema
}

func (CoverLetterVersion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (CoverLetterVersion) Fields() []ent.Field {
	return []ent.Field{
		field.Int("version_number").
			Comment("Version number"),
		field.Text("content").
			NotEmpty().
			Comment("Version content"),
		field.Int("char_count").
			Optional().
			Nillable().
			Comment("Character count"),
		field.Text("change_summary").
			Optional().
			Comment("Change summary"),
		field.JSON("scores", map[string]interface{}{}).
			Optional().
			Comment("Scores: {specificity, jobFit, companyFit, authenticity}"),
	}
}

func (CoverLetterVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cover_letter", CoverLetter.Type).
			Ref("versions").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("coaching_session", CoachingSession.Type).
			Ref("cover_letter_versions").
			Unique(),
	}
}

func (CoverLetterVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version_number"),
	}
}
