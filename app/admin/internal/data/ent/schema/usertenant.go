package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UserTenant struct{ ent.Schema }

func (UserTenant) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").Comment("SysUser ID"),
		field.Int64("tenant_id").Comment("Tenant ID"),
		field.JSON("role_labels", []string{}).Optional(),
	}
}

func (UserTenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("user_tenants").Required().Unique().Field("user_id"),
		edge.From("tenant", Tenant.Type).Ref("user_tenants").Required().Unique().Field("tenant_id"),
	}
}

func (UserTenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "tenant_id").Unique(), // 防重复
		index.Fields("tenant_id"),
		index.Fields("user_id"),
	}
}

func (UserTenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
