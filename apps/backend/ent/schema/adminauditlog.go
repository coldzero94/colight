package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type AdminAuditLog struct {
	ent.Schema
}

func (AdminAuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (AdminAuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("admin_id", uuid.UUID{}).
			Comment("Admin who performed the action"),
		field.String("action").
			MaxLen(50).
			Comment("Action performed: role_change, config_update, suspend, unsuspend, prompt_update"),
		field.String("target_type").
			MaxLen(50).
			Comment("Target entity type: user, config, prompt"),
		field.String("target_id").
			Optional().
			MaxLen(255).
			Comment("Target entity ID"),
		field.Text("old_value").
			Optional().
			Nillable().
			Comment("Previous value (JSON)"),
		field.Text("new_value").
			Optional().
			Nillable().
			Comment("New value (JSON)"),
		field.String("ip_address").
			Optional().
			MaxLen(45).
			Comment("Client IP address"),
	}
}

func (AdminAuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("admin_id"),
		index.Fields("action"),
		index.Fields("created_at"),
	}
}
