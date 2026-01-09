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

// UserAccounts holds the schema definition for the UserAccounts entity.
type UserAccount struct {
	ent.Schema
}

// Fields of the UserAccount.
func (UserAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").Positive().Comment("Reference to SysUser ID"),
		field.String("platform").MaxLen(32).Comment("Social media platform (e.g., Twitter, Facebook)"),
		field.String("identifier").MaxLen(255).Comment("User's account identifier on the platform"),
		field.String("name").Nillable().Comment("User's name on the platform"),
		field.Time("created_at").Default(time.Now).Comment("Record creation timestamp"),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Comment("Record last update timestamp"),
		field.Time("deleted_at").Optional().Nillable().Comment("Soft delete flag, null if not deleted"),
	}
}

// Edges of the UserAccounts.
func (UserAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("accounts").Required().Unique().Field("user_id"),
	}
}

func (UserAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}

func (UserAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),                         // 单字段索引
		index.Fields("platform", "identifier").Unique(), // 若希望平台+账号唯一，可添加 Unique
	}
}
