// admin/common/authz/subject.go
package authz

import (
	"context"

	"github.com/yc-alpha/admin/ent"
	"github.com/yc-alpha/admin/ent/userrole"
)

// Subject ABAC主体，包含用户的所有授权相关属性
type Subject struct {
	UserID     int64    `json:"user_id"`
	Username   string   `json:"username"`
	TenantID   int64    `json:"tenant_id"`   // 当前操作的租户
	RoleCodes  []string `json:"role_codes"`  // 用户在当前租户下的角色code列表
	IsPlatform bool     `json:"is_platform"` // 是否拥有平台级角色
}

// HasRole 检查是否拥有某个角色
func (s *Subject) HasRole(roleCode string) bool {
	for _, code := range s.RoleCodes {
		if code == roleCode {
			return true
		}
	}
	return false
}

// HasAnyRole 检查是否拥有任意一个角色
func (s *Subject) HasAnyRole(roleCodes ...string) bool {
	for _, code := range roleCodes {
		if s.HasRole(code) {
			return true
		}
	}
	return false
}

// SubjectBuilder 从数据库构建Subject
type SubjectBuilder struct {
	client *ent.Client
}

func NewSubjectBuilder(client *ent.Client) *SubjectBuilder {
	return &SubjectBuilder{client: client}
}

// BuildSubject 根据userID和tenantID构建Subject
func (b *SubjectBuilder) BuildSubject(ctx context.Context, userID int64, tenantID int64) (*Subject, error) {
	// 查询用户基本信息
	user, err := b.client.User.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	sub := &Subject{
		UserID:     userID,
		Username:   user.Username,
		TenantID:   tenantID,
		RoleCodes:  make([]string, 0),
		IsPlatform: false,
	}

	// 查询用户的角色
	// 1. 平台级角色（tenant_id IS NULL）
	platformRoles, err := b.client.UserRole.Query().
		Where(
			userrole.UserIDEQ(userID),
			userrole.TenantIDIsNil(),
		).
		WithRole().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, ur := range platformRoles {
		if ur.Edges.Role != nil && ur.Edges.Role.IsActive {
			sub.RoleCodes = append(sub.RoleCodes, ur.Edges.Role.Code)
			sub.IsPlatform = true
		}
	}

	// 2. 租户级角色（当前租户）
	if tenantID > 0 {
		tenantRoles, err := b.client.UserRole.Query().
			Where(
				userrole.UserIDEQ(userID),
				userrole.TenantIDEQ(tenantID),
			).
			WithRole().
			All(ctx)
		if err != nil {
			return nil, err
		}

		for _, ur := range tenantRoles {
			if ur.Edges.Role != nil && ur.Edges.Role.IsActive {
				sub.RoleCodes = append(sub.RoleCodes, ur.Edges.Role.Code)
			}
		}
	}

	return sub, nil
}
