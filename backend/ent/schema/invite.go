package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Invite holds the schema definition for the Invite entity.
type Invite struct {
	ent.Schema
}

// Fields of the Invite.
func (Invite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("token").
			Unique().
			NotEmpty(),
		field.String("email").
			NotEmpty(),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("role").
			Values("owner", "admin", "member").
			Default("member"),
		field.Enum("project_permission").
			Values("edit", "view").
			Default("view").
			Optional().
			Nillable(),
		field.UUID("invited_by_id", uuid.UUID{}),
		field.Time("expires_at"),
		field.Time("used_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Invite.
func (Invite) Edges() []ent.Edge {
	return []ent.Edge{
		// Invite belongs to an organization
		edge.From("organization", Organization.Type).
			Ref("invites").
			Field("organization_id").
			Unique().
			Required(),
		// Invite optionally belongs to a project
		edge.From("project", Project.Type).
			Ref("invites").
			Field("project_id").
			Unique(),
		// Invite was sent by a user
		edge.From("invited_by", User.Type).
			Ref("sent_invites").
			Field("invited_by_id").
			Unique().
			Required(),
	}
}

// Indexes of the Invite.
func (Invite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token").Unique(),
		index.Fields("email", "organization_id"),
	}
}

