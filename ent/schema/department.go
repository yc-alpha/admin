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

type Department struct{ ent.Schema }

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable().DefaultFunc(snowflake.GenId).Comment("Primary Key ID"),
		field.Int64("tenant_id").Comment("Tenant ID"),
		field.Int64("parent_id").Comment("Parent Department ID"),
		field.String("name").Comment("Name of the department"),
		field.String("path").SchemaType(map[string]string{"postgres": "ltree"}).Comment("save ltree path"),
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

func (Department) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"), // 多租户条件
		index.Fields("tenant_id", "parent_id").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")), // 软删除过滤
		index.Fields("created_at"), // 创建时间排序
		index.Fields("path").
			Annotations(entsql.IndexAnnotation{
				Types: map[string]string{
					"postgres": "GIST", // GIST 更适合 ltree
				},
			}),
	}
}

func (Department) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
