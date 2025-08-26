package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}
