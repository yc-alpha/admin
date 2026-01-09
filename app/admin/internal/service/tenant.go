package service

import (
	"context"
	"fmt"

	"github.com/yc-alpha/admin/ent"
	"github.com/yc-alpha/admin/ent/tenant"
	"github.com/yc-alpha/logger"
)

// TenantService 租户管理服务
type TenantService struct {
	client *ent.Client
}

// NewTenantService 创建租户服务
func NewTenantService(client *ent.Client) *TenantService {
	return &TenantService{
		client: client,
	}
}

// TenantType 租户类型
type TenantType string

const (
	TenantTypeNormal TenantType = "NORMAL" // 普通租户
	TenantTypeGroup  TenantType = "GROUP"  // 集团型租户
	TenantTypeSub    TenantType = "SUB"    // 子租户
)

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name       string         `json:"name"`
	OwnerID    int64          `json:"owner_id"`
	Type       TenantType     `json:"type"`
	ParentID   *int64         `json:"parent_id,omitempty"`
	Status     string         `json:"status,omitempty"`
	ExpiredAt  *string        `json:"expired_at,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	CreatedBy  *int64         `json:"created_by,omitempty"`
}

// CreateTenantResponse 创建租户响应
type CreateTenantResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Tenant  *ent.Tenant `json:"tenant,omitempty"`
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*CreateTenantResponse, error) {
	// 验证租户类型约束
	if err := s.validateTenantType(req.Type, req.ParentID); err != nil {
		return &CreateTenantResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 如果是子租户，验证父租户是否存在且为集团型租户
	if req.Type == TenantTypeSub && req.ParentID != nil {
		parent, err := s.client.Tenant.Get(ctx, *req.ParentID)
		if err != nil {
			return &CreateTenantResponse{
				Success: false,
				Message: "父租户不存在",
			}, nil
		}
		if parent.Type != tenant.TypeGROUP {
			return &CreateTenantResponse{
				Success: false,
				Message: "只有集团型租户才能创建子租户",
			}, nil
		}
	}

	// 创建租户
	tenantBuilder := s.client.Tenant.Create().
		SetName(req.Name).
		SetOwnerID(req.OwnerID).
		SetType(tenant.Type(req.Type))

	if req.ParentID != nil {
		tenantBuilder.SetParentID(*req.ParentID)
	}

	if req.Status != "" {
		tenantBuilder.SetStatus(tenant.Status(req.Status))
	}

	if req.Attributes != nil {
		tenantBuilder.SetAttributes(req.Attributes)
	}

	if req.CreatedBy != nil {
		tenantBuilder.SetCreatedBy(*req.CreatedBy)
	}

	createdTenant, err := tenantBuilder.Save(ctx)
	if err != nil {
		logger.Errorf("创建租户失败: %v", err)
		return &CreateTenantResponse{
			Success: false,
			Message: "创建租户失败",
		}, nil
	}

	logger.Infof("成功创建租户: %s (ID: %d)", createdTenant.Name, createdTenant.ID)
	return &CreateTenantResponse{
		Success: true,
		Message: "租户创建成功",
		Tenant:  createdTenant,
	}, nil
}

// GetTenantHierarchy 获取租户层级结构
func (s *TenantService) GetTenantHierarchy(ctx context.Context, tenantID int64) (*ent.Tenant, error) {
	return s.client.Tenant.Query().
		Where(tenant.ID(tenantID)).
		WithParent().
		WithChildren().
		Only(ctx)
}

// GetRootTenants 获取根租户列表
func (s *TenantService) GetRootTenants(ctx context.Context) ([]*ent.Tenant, error) {
	return s.client.Tenant.Query().
		Where(
			tenant.ParentIDIsNil(),
			tenant.DeletedAtIsNil(),
		).
		WithChildren().
		Order(ent.Asc(tenant.FieldCreatedAt)).
		All(ctx)
}

// GetSubTenants 获取子租户列表
func (s *TenantService) GetSubTenants(ctx context.Context, parentID int64) ([]*ent.Tenant, error) {
	return s.client.Tenant.Query().
		Where(
			tenant.ParentID(parentID),
			tenant.DeletedAtIsNil(),
		).
		WithParent().
		Order(ent.Asc(tenant.FieldCreatedAt)).
		All(ctx)
}

// GetGroupTenants 获取集团型租户列表
func (s *TenantService) GetGroupTenants(ctx context.Context) ([]*ent.Tenant, error) {
	return s.client.Tenant.Query().
		Where(
			tenant.TypeEQ(tenant.TypeGROUP),
			tenant.DeletedAtIsNil(),
		).
		WithChildren().
		Order(ent.Asc(tenant.FieldCreatedAt)).
		All(ctx)
}

// GetTenantPath 获取租户完整路径
func (s *TenantService) GetTenantPath(ctx context.Context, tenantID int64) ([]*ent.Tenant, error) {
	var path []*ent.Tenant

	current, err := s.client.Tenant.Get(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	path = append(path, current)

	// 向上遍历到根租户
	for current.ParentID != nil {
		parent, err := s.client.Tenant.Get(ctx, *current.ParentID)
		if err != nil {
			return nil, err
		}
		path = append([]*ent.Tenant{parent}, path...)
		current = parent
	}

	return path, nil
}

// validateTenantType 验证租户类型约束
func (s *TenantService) validateTenantType(tenantType TenantType, parentID *int64) error {
	switch tenantType {
	case TenantTypeNormal:
		if parentID != nil {
			return fmt.Errorf("普通租户不能有父租户")
		}
	case TenantTypeGroup:
		if parentID != nil {
			return fmt.Errorf("集团型租户不能有父租户")
		}
	case TenantTypeSub:
		if parentID == nil {
			return fmt.Errorf("子租户必须指定父租户")
		}
	default:
		return fmt.Errorf("无效的租户类型: %s", tenantType)
	}
	return nil
}

// CanCreateSubTenant 检查是否可以创建子租户
func (s *TenantService) CanCreateSubTenant(ctx context.Context, tenantID int64) (bool, error) {
	t, err := s.client.Tenant.Get(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 只有集团型租户才能创建子租户
	return t.Type == tenant.TypeGROUP, nil
}

// GetTenantStatistics 获取租户统计信息
func (s *TenantService) GetTenantStatistics(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// 统计各类型租户数量
	normalCount, err := s.client.Tenant.Query().
		Where(tenant.TypeEQ(tenant.TypeNORMAL), tenant.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["normal"] = normalCount

	groupCount, err := s.client.Tenant.Query().
		Where(tenant.TypeEQ(tenant.TypeGROUP), tenant.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["group"] = groupCount

	subCount, err := s.client.Tenant.Query().
		Where(tenant.TypeEQ(tenant.TypeSUB), tenant.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["sub"] = subCount

	// 统计总租户数
	totalCount, err := s.client.Tenant.Query().
		Where(tenant.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["total"] = totalCount

	return stats, nil
}
