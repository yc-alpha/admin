// ent/schema/casbin_rule.go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CasbinRule struct {
	ent.Schema
}

func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("ptype").MaxLen(100).Comment("策略类型：p或g"),
		field.String("v0").MaxLen(100).Default(""),
		field.String("v1").MaxLen(100).Default(""),
		field.String("v2").MaxLen(100).Default(""),
		field.String("v3").MaxLen(100).Default(""),
		field.String("v4").MaxLen(100).Default(""),
		field.String("v5").MaxLen(100).Default(""),
	}
}

func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ptype", "v0", "v1", "v2", "v3", "v4", "v5"),
	}
}
