// ent/schema/role.go
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

type Role struct {
	ent.Schema
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("code").NotEmpty().Comment("角色编码，如 admin, tenant_admin, sales_manager"),
		field.String("name").NotEmpty().Comment("角色名称"),
		field.Int64("tenant_id").Optional().Nillable().Comment("租户ID"),
		field.Bool("is_system").Default(false).Comment("是否系统预置角色（系统预置的不可删除/修改code）"),
		field.String("description").Optional(),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_roles", UserRole.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("tenant", Tenant.Type).Ref("roles").Unique().Field("tenant_id"),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		// 平台级角色code全局唯一
		// 租户内角色code租户内唯一
		index.Fields("tenant_id", "code").Unique(),
	}
}

func (Role) Annotations() []schema.Annotation {
	// 把 WithComments 放这里，ent 在迁移时会把 field.Comment 写入 DB
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
