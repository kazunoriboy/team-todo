package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("password_hash").
			NotEmpty().
			Sensitive(),
		field.String("display_name").
			NotEmpty(),
		field.UUID("last_org_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("last_project_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// User belongs to many organizations through OrganizationMember
		edge.To("organizations", Organization.Type).
			Through("organization_memberships", OrganizationMember.Type),
		// User belongs to many projects through ProjectMember
		edge.To("projects", Project.Type).
			Through("project_memberships", ProjectMember.Type),
		// User sent invites
		edge.To("sent_invites", Invite.Type),
		// Last accessed organization
		edge.To("last_organization", Organization.Type).
			Unique().
			Field("last_org_id"),
		// Last accessed project
		edge.To("last_project", Project.Type).
			Unique().
			Field("last_project_id"),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
	}
}

