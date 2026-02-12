package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ExperienceWeapon struct {
	ent.Schema
}

func (ExperienceWeapon) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (ExperienceWeapon) Fields() []ent.Field {
	return []ent.Field{
		field.String("weapon_code").
			NotEmpty().
			MaxLen(10).
			Comment("Weapon category code (W01, W01-A, etc.)"),
		field.Float("confidence").
			Default(0).
			Comment("AI classification confidence (0~1)"),
		field.Bool("is_primary").
			Default(false).
			Comment("Whether this is the primary weapon"),
		field.Text("reasoning").
			Optional().
			Comment("AI classification reasoning"),
		field.Bool("user_confirmed").
			Default(false).
			Comment("Whether user confirmed the classification"),
		field.Bool("user_modified").
			Default(false).
			Comment("Whether user modified the classification"),
	}
}

func (ExperienceWeapon) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("experience", Experience.Type).
			Ref("weapons").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		// Note: weapon_code references WeaponCategory.code (string)
		// This is handled at the application level via the weapon_code field
	}
}

func (ExperienceWeapon) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("weapon_code"),
		index.Fields("is_primary"),
	}
}
