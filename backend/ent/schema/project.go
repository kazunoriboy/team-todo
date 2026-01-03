package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}),
		field.String("name").
			NotEmpty(),
		field.Bool("is_private").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		// Project belongs to an organization
		edge.From("organization", Organization.Type).
			Ref("projects").
			Field("organization_id").
			Unique().
			Required(),
		// Project has many members through ProjectMember
		edge.From("members", User.Type).
			Ref("projects").
			Through("project_memberships", ProjectMember.Type),
		// Project has many invites
		edge.To("invites", Invite.Type),
		// Users who last accessed this project
		edge.From("last_accessed_by", User.Type).
			Ref("last_project"),
	}
}

// Indexes of the Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "name"),
	}
}

