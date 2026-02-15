package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type SystemConfig struct {
	ent.Schema
}

func (SystemConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (SystemConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("config_key").
			MaxLen(100).
			Unique().
			Comment("Configuration key"),
		field.Text("config_value").
			Comment("Configuration value (encrypted if is_secret)"),
		field.String("description").
			Optional().
			MaxLen(500).
			Comment("Description of this config"),
		field.String("category").
			MaxLen(50).
			Comment("Category: api_key, model, limit, cost"),
		field.Bool("is_secret").
			Default(false).
			Comment("Whether value is encrypted"),
		field.UUID("updated_by", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Admin who last updated this config"),
	}
}

func (SystemConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category"),
		index.Fields("config_key").Unique(),
	}
}
