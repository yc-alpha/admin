// ent/schema/user_role.go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UserRole struct {
	ent.Schema
}

func (UserRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id").Comment("用户ID"),
		field.Int64("role_id").Comment("角色ID"),
		field.Int64("tenant_id").Optional().Nillable().Comment("租户ID"),
		field.Time("granted_at").Default(time.Now).Immutable(),
		field.Time("expires_at").Optional().Nillable().Comment("过期时间,NULL表示永久有效"),
	}
}

func (UserRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("user_roles").Required().Unique().Field("user_id"),
		edge.From("role", Role.Type).Ref("user_roles").Required().Unique().Field("role_id"),
		edge.From("tenant", Tenant.Type).Ref("user_roles").Unique().Field("tenant_id"),
	}
}

func (UserRole) Indexes() []ent.Index {
	return []ent.Index{
		// 用户在同一租户下不能重复拥有同一角色
		index.Fields("user_id", "role_id", "tenant_id").Unique(),
		index.Fields("user_id"),
		index.Fields("role_id"),
		index.Fields("tenant_id"),
	}
}

func (UserRole) Annotations() []schema.Annotation {
	// 把 WithComments 放这里，ent 在迁移时会把 field.Comment 写入 DB
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
