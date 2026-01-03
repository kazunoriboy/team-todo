package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OrganizationMember holds the schema definition for the OrganizationMember entity.
// This is a join table between User and Organization with additional role field.
type OrganizationMember struct {
	ent.Schema
}

// Annotations of the OrganizationMember.
func (OrganizationMember) Annotations() []schema.Annotation {
	return nil
}

// Fields of the OrganizationMember.
func (OrganizationMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("organization_id", uuid.UUID{}),
		field.Enum("role").
			Values("owner", "admin", "member").
			Default("member"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the OrganizationMember.
func (OrganizationMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("organization", Organization.Type).
			Unique().
			Required().
			Field("organization_id"),
	}
}

// Indexes of the OrganizationMember.
func (OrganizationMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").Unique(),
	}
}

