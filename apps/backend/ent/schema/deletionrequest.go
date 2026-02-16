package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DeletionRequest tracks user data deletion requests for PIPA compliance.
type DeletionRequest struct {
	ent.Schema
}

func (DeletionRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

func (DeletionRequest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).
			Comment("References user_profiles(id)"),
		field.Text("reason").
			Optional().
			Comment("Reason for deletion request"),
		field.Time("scheduled_at").
			Comment("When the deletion will be executed (30 days from request)"),
		field.Enum("status").
			Values("pending", "cancelled", "completed").
			Default("pending").
			Comment("Deletion request status"),
		field.UUID("requested_by", uuid.UUID{}).
			Comment("Admin who initiated the request"),
		field.Time("cancelled_at").
			Optional().
			Nillable().
			Comment("When the request was cancelled"),
	}
}

func (DeletionRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", UserProfile.Type).
			Ref("deletion_requests").
			Unique().
			Required().
			Field("user_id").
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (DeletionRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
	}
}
