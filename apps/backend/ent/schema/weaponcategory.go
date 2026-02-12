package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type WeaponCategory struct {
	ent.Schema
}

func (WeaponCategory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (WeaponCategory) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Unique().
			NotEmpty().
			MaxLen(10).
			Comment("Weapon code: W01, W01-A, etc."),
		field.String("parent_code").
			Optional().
			MaxLen(10).
			Comment("Parent weapon code (NULL for top-level)"),
		field.String("name").
			NotEmpty().
			MaxLen(50).
			Comment("Weapon name"),
		field.Text("description").
			Optional().
			Comment("Weapon description"),
		field.JSON("keywords", []string{}).
			Optional().
			Comment("Related keywords array"),
		field.JSON("question_patterns", []string{}).
			Optional().
			Comment("Frequently asked question patterns"),
		field.Int("display_order").
			Default(0).
			Comment("Display order"),
		field.String("icon").
			Optional().
			MaxLen(10).
			Comment("Emoji icon"),
		field.String("color").
			Optional().
			MaxLen(7).
			Comment("HEX color code"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether this category is active"),
	}
}

func (WeaponCategory) Edges() []ent.Edge {
	return []ent.Edge{
		// Linked experience weapons
		edge.To("experience_weapons", ExperienceWeapon.Type),
		// Note: Self-referential parent-children via parent_code (string)
		// is handled at the application level, not as an Ent edge
	}
}

func (WeaponCategory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_code"),
		index.Fields("code"),
	}
}
