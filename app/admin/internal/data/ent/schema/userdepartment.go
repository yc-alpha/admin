package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type UserDepartment struct{ ent.Schema }

func (UserDepartment) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable(),
		field.Int64("user_id").Comment("SysUser ID"),
		field.Int64("dept_id").Comment("Department ID"),
		field.Int64("tenant_id").Comment("Tenant ID"),
		field.JSON("attributes", map[string]any{}).Default(map[string]any{}),
	}
}

func (UserDepartment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("user_departments").Required().Unique().Field("user_id"),
		edge.From("department", Department.Type).Ref("user_departments").Required().Unique().Field("dept_id"),
	}
}

func (UserDepartment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
