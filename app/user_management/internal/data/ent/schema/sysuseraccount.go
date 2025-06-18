package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SysUserAccounts holds the schema definition for the SysUserAccounts entity.
type SysUserAccount struct {
	ent.Schema
}

// Fields of the SysUserAccount.
func (SysUserAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").Positive().Comment("Reference to SysUser ID"),
		field.String("platform").Comment("Social media platform (e.g., Twitter, Facebook)"),
		field.String("identifier").Comment("User's account identifier on the platform"),
		field.String("name").Nillable().Comment("User's name on the platform"),
		field.Time("created_at").Default(time.Now).Comment("Record creation timestamp"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("Record last update timestamp"),
		field.Time("deleted_at").Optional().Nillable().Comment("Soft delete flag, null if not deleted"),
	}
}

// Edges of the SysUserAccounts.
func (SysUserAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", SysUser.Type).
			Ref("accounts").
			Required().
			Unique().
			Field("user_id").
			Annotations(
				entsql.WithComments(true),
			),
	}
}

func (SysUserAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),                         // 单字段索引
		index.Fields("platform", "identifier").Unique(), // 若希望平台+账号唯一，可添加 Unique
	}
}
