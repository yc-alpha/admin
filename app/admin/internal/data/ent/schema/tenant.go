package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/yc-alpha/admin/common/snowflake"
)

type Tenant struct{ ent.Schema }

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable().DefaultFunc(snowflake.GenId).Comment("Primary Key ID"),
		field.String("name").MaxLen(128).NotEmpty().Comment("Name of the tenant"),
		field.Int64("owner_id").Comment("Owner user ID of the tenant"),
		field.Enum("status").Values("ACTIVE", "DISABLED", "EXPIRED", "PENDING").Default("PENDING").Comment("Status of the tenant"),
		field.Time("expired_at").Optional().Nillable().Comment("Expiration time of the tenant"),
		field.JSON("attributes", map[string]any{}).Default(map[string]any{}).Comment(""),
		field.Int64("created_by").Optional().Nillable().Comment("User who created this record"),
		field.Int64("updated_by").Optional().Nillable().Comment("User who last updated this record"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("Creation timestamp of this record"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("Last update timestamp of this record"),
		field.Time("deleted_at").Optional().Nillable().Comment("Timestamp when the record was deleted, if applicable"),
	}
}

func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user_tenants", UserTenant.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("departments", Department.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		// 根据 owner_id 查询租户
		index.Fields("owner_id"),
		// 有效租户（未删除）
		index.Fields("status").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		// 到期时间查询（批量清理/检查即将过期的租户）
		index.Fields("expired_at"),
		// 创建时间排序（最新租户）
		index.Fields("created_at"),
	}
}

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
