package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Department struct{ ent.Schema }

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable(),
		field.Int64("tenant_id").Comment("Tenant ID"),
		field.Int64("parent_id").Comment("Parent Department ID"),
		field.String("name").Comment("Name of the department"),
		field.String("path").Comment(""), // 保存 ltree path 文本
		field.JSON("attributes", map[string]any{}).Default(map[string]any{}),
		field.Int64("created_by").Optional().Nillable().Comment("User who created this record"),
		field.Int64("updated_by").Optional().Nillable().Comment("User who last updated this record"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("Creation timestamp of this record"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("Last update timestamp of this record"),
		field.Time("deleted_at").Optional().Nillable().Comment("Timestamp when the record was deleted, if applicable"),
	}
}

func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("departments").Required().Unique().Field("tenant_id"),
		edge.To("user_departments", UserDepartment.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Department) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
