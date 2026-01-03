package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ProjectMember holds the schema definition for the ProjectMember entity.
// This is a join table between User and Project with additional permission field.
type ProjectMember struct {
	ent.Schema
}

// Fields of the ProjectMember.
func (ProjectMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("project_id", uuid.UUID{}),
		field.Enum("permission").
			Values("edit", "view").
			Default("view"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ProjectMember.
func (ProjectMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("project", Project.Type).
			Unique().
			Required().
			Field("project_id"),
	}
}

// Indexes of the ProjectMember.
func (ProjectMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "project_id").Unique(),
	}
}

