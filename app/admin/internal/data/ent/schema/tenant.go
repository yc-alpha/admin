package schema

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/yc-alpha/admin/app/admin/internal/constant"
	gen "github.com/yc-alpha/admin/app/admin/internal/data/ent"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/hook"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/tenant"
	"github.com/yc-alpha/admin/common/snowflake"
)

type Tenant struct{ ent.Schema }

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable().DefaultFunc(snowflake.GenId).Comment("Primary Key ID"),
		field.String("name").MaxLen(128).NotEmpty().Comment("Name of the tenant"),
		field.Int64("owner_id").Comment("Owner user ID of the tenant"),
		field.Enum("type").Values("ROOT", "NORMAL", "GROUP", "SUB").Default("NORMAL").Comment("Tenant type: ROOT(系统租户), NORMAL(普通租户), GROUP(集团型租户), SUB(子租户)"),
		field.Int64("parent_id").Optional().Nillable().Comment("Parent tenant ID for sub-tenants"),
		field.String("path").SchemaType(map[string]string{"postgres": "ltree"}).Optional().Nillable().Comment("Tenant hierarchy path using ltree"),
		field.Int32("level").Default(0).Comment("Tenant hierarchy level (0 for root tenants)"),
		field.Enum("status").Values("ACTIVE", "DISABLED", "EXPIRED", "PENDING").Default("PENDING").Comment("Status of the tenant"),
		field.Time("expired_at").Optional().Nillable().Comment("Expiration time of the tenant"),
		field.JSON("attributes", map[string]any{}).Default(map[string]any{}).Comment("Tenant attributes and metadata"),
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
		// 父子租户关系
		edge.From("parent", Tenant.Type).Ref("children").Unique().Field("parent_id"),
		edge.To("children", Tenant.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		// 根据 owner_id 查询租户
		index.Fields("owner_id"),
		// 租户类型查询
		index.Fields("type"),
		// 父子租户关系查询
		index.Fields("parent_id"),
		index.Fields("type", "parent_id").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		// 租户层级查询
		index.Fields("level"),
		index.Fields("path").Annotations(entsql.IndexAnnotation{
			Types: map[string]string{
				"postgres": "GIST", // GIST 更适合 ltree
			},
		}),
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
		// 添加约束：只有集团型租户才能有子租户
		entsql.Checks(map[string]string{
			"tenant_type_check": fmt.Sprintf(`
				(type = 'ROOT' AND parent_id IS NULL) OR
				(type IN ('GROUP','NORMAL') AND parent_id = %d) OR 
				(type = 'SUB' AND parent_id IS NOT NULL)`, constant.ROOT_TENANT_ID),
		}),
	}
}

// Hooks 处理租户层级关系
func (Tenant) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(func(next ent.Mutator) ent.Mutator {
			return hook.TenantFunc(func(ctx context.Context, m *gen.TenantMutation) (ent.Value, error) {
				// 此时m已经包含了id，无论是用户设置还是雪花算法自动生成的
				tenantTypeVal, ok := m.Field("type")
				if !ok {
					return nil, fmt.Errorf("缺少租户类型字段")
				}
				tenantType, ok := tenantTypeVal.(tenant.Type)
				if !ok {
					return nil, fmt.Errorf("租户类型字段类型不正确: %T", tenantTypeVal)
				}

				// ROOT限制
				if tenantType == tenant.TypeROOT {
					count, err := m.Client().Tenant.Query().Where(tenant.TypeEQ(tenant.TypeROOT)).Count(ctx)
					if err != nil {
						return nil, err
					}
					if count > 0 {
						return nil, fmt.Errorf("禁止新建系统租户(ROOT)")
					}
					m.SetField("level", 0)
					m.SetField("path", fmt.Sprintf("%d", constant.ROOT_TENANT_ID))
					m.SetField("parent_id", nil)
					return next.Mutate(ctx, m)
				}

				// 校验父租户存在
				parentID, _ := m.Field("parent_id")
				if parentID == nil {
					return nil, fmt.Errorf("必须指定父租户")
				}
				pid, ok := parentID.(int64)
				if !ok {
					return nil, fmt.Errorf("parent_id 类型不正确: %T", parentID)
				}
				parent, err := m.Client().Tenant.Get(ctx, pid)
				if err != nil {
					return nil, fmt.Errorf("获取父租户失败: %w", err)
				}
				// 逻辑层级验证
				switch tenantType {
				case tenant.TypeGROUP, tenant.TypeNORMAL:
					if parent.Type != tenant.TypeROOT {
						return nil, fmt.Errorf("无效的父租户类型")
					}
				case tenant.TypeSUB:
					if parent.Type != tenant.TypeGROUP {
						return nil, fmt.Errorf("子租户的父租户必须是集团型租户")
					}
				default:
					return nil, fmt.Errorf("未知租户类型: %v", tenantType)
				}

				// 计算层级
				m.SetField("level", parent.Level+1)
				// 设置路径
				id, _ := m.ID()
				m.SetField("path", fmt.Sprintf("%s.%d", *parent.Path, id))
				return next.Mutate(ctx, m)
			})
		}, ent.OpCreate),

		// 删除前禁止删除 ROOT 租户
		hook.On(func(next ent.Mutator) ent.Mutator {
			return hook.TenantFunc(func(ctx context.Context, m *gen.TenantMutation) (ent.Value, error) {
				id, exist := m.ID()
				if !exist {
					return nil, fmt.Errorf("删除操作缺少租户 ID")
				}
				t, err := m.Client().Tenant.Get(ctx, id)
				if err != nil {
					return nil, fmt.Errorf("获取租户失败: %w", err)
				}
				if t.Type == tenant.TypeROOT {
					return nil, fmt.Errorf("系统租户(ROOT)禁止删除")
				}
				return next.Mutate(ctx, m)
			})
		}, ent.OpDelete|ent.OpDeleteOne),

		// 修改前保护禁止修改的字段
		hook.On(func(next ent.Mutator) ent.Mutator {
			return hook.TenantFunc(func(ctx context.Context, m *gen.TenantMutation) (ent.Value, error) {
				id, ok := m.ID()
				if !ok {
					return nil, fmt.Errorf("更新操作缺少租户 ID")
				}
				t, err := m.Client().Tenant.Get(ctx, id)
				if err != nil {
					return nil, fmt.Errorf("获取租户失败: %w", err)
				}

				// 禁止修改 ROOT
				if t.Type == tenant.TypeROOT {
					return nil, fmt.Errorf("系统租户(ROOT)禁止修改")
				}

				// 禁止修改层级字段
				protected := []string{"level", "parent_id", "type", "path"}
				for _, f := range protected {
					if _, ok := m.Field(f); ok {
						return nil, fmt.Errorf("禁止修改字段: %s", f)
					}
				}

				return next.Mutate(ctx, m)
			})
		}, ent.OpUpdate|ent.OpUpdateOne),
	}
}
