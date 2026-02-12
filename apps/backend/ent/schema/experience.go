package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Experience struct {
	ent.Schema
}

func (Experience) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Experience) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References auth.users(id)"),
		field.String("title").
			NotEmpty().
			MaxLen(200).
			Comment("Experience title"),
		field.String("category").
			Optional().
			MaxLen(50).
			Comment("Activity type: 인턴, 대외활동, 프로젝트, 아르바이트"),
		field.Time("period_start").
			Optional().
			Nillable().
			Comment("Start date"),
		field.Time("period_end").
			Optional().
			Nillable().
			Comment("End date"),
		field.String("role").
			Optional().
			MaxLen(100).
			Comment("Role in the experience"),
		field.Text("content").
			NotEmpty().
			Comment("Detailed experience content"),
		field.Text("result").
			Optional().
			Comment("Outcome / achievements"),
		field.Text("star_situation").
			Optional().
			Comment("STAR: Situation"),
		field.Text("star_task").
			Optional().
			Comment("STAR: Task"),
		field.Text("star_action").
			Optional().
			Comment("STAR: Action"),
		field.Text("star_result").
			Optional().
			Comment("STAR: Result"),
		field.JSON("keywords", []string{}).
			Optional().
			Comment("Core keywords array"),
		// embedding VECTOR(1536) — handled via raw SQL migration
		field.String("source").
			Default("manual").
			MaxLen(20).
			Comment("Input method: manual, interview"),
		field.Bool("is_archived").
			Default(false).
			Comment("Whether experience is archived"),
	}
}

func (Experience) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("experiences").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("tags", ExperienceTag.Type),
		edge.To("weapons", ExperienceWeapon.Type),
		edge.To("usages", ExperienceUsage.Type),
	}
}

func (Experience) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "category"),
		index.Fields("user_id", "created_at"),
		index.Fields("user_id", "is_archived"),
	}
}
