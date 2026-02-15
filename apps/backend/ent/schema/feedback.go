package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Feedback stores user feedback submissions.
type Feedback struct {
	ent.Schema
}

func (Feedback) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (Feedback) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References user_profiles(id)"),
		field.Enum("category").
			Values("bug", "improvement", "other").
			Comment("Feedback category"),
		field.Text("content").
			NotEmpty().
			Comment("Feedback content"),
		field.String("page_url").
			Optional().
			MaxLen(500).
			Comment("Page URL where feedback was submitted"),
		field.String("user_agent").
			Optional().
			MaxLen(500).
			Comment("Browser user agent string"),
	}
}

func (Feedback) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("feedbacks").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Feedback) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
	}
}
