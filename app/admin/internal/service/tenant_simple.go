package service

import (
	"context"
	"fmt"

	"github.com/yc-alpha/admin/app/admin/internal/data/ent"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/tenant"
	"github.com/yc-alpha/logger"
)

// SimpleTenantService 简化的租户服务
type SimpleTenantService struct {
	client *ent.Client
}

// NewSimpleTenantService 创建简化租户服务
func NewSimpleTenantService(client *ent.Client) *SimpleTenantService {
	return &SimpleTenantService{
		client: client,
	}
}

// CreateTenant 创建租户
func (s *SimpleTenantService) CreateTenant(ctx context.Context, name string, ownerID int64, tenantType string) (*ent.Tenant, error) {
	// 创建租户
	tenantBuilder := s.client.Tenant.Create().
		SetName(name).
		SetOwnerID(ownerID)

	// 设置租户类型
	switch tenantType {
	case "NORMAL":
		tenantBuilder.SetType(tenant.TypeNORMAL)
	case "GROUP":
		tenantBuilder.SetType(tenant.TypeGROUP)
	case "SUB":
		tenantBuilder.SetType(tenant.TypeSUB)
	default:
		return nil, fmt.Errorf("无效的租户类型: %s", tenantType)
	}

	// 设置默认状态
	tenantBuilder.SetStatus(tenant.StatusACTIVE)

	createdTenant, err := tenantBuilder.Save(ctx)
	if err != nil {
		logger.Errorf("创建租户失败: %v", err)
		return nil, fmt.Errorf("创建租户失败: %w", err)
	}

	logger.Infof("成功创建租户: %s (ID: %d, Type: %s)", createdTenant.Name, createdTenant.ID, createdTenant.Type)
	return createdTenant, nil
}

// GetRootTenants 获取根租户列表
func (s *SimpleTenantService) GetRootTenants(ctx context.Context) ([]*ent.Tenant, error) {
	return s.client.Tenant.Query().
		Where(tenant.ParentIDIsNil()).
		Order(ent.Asc(tenant.FieldCreatedAt)).
		All(ctx)
}

// GetSubTenants 获取子租户列表
func (s *SimpleTenantService) GetSubTenants(ctx context.Context, parentID int64) ([]*ent.Tenant, error) {
	return s.client.Tenant.Query().
		Where(tenant.ParentID(parentID)).
		Order(ent.Asc(tenant.FieldCreatedAt)).
		All(ctx)
}

// GetTenantByID 根据ID获取租户
func (s *SimpleTenantService) GetTenantByID(ctx context.Context, tenantID int64) (*ent.Tenant, error) {
	return s.client.Tenant.Get(ctx, tenantID)
}

// GetTenantStatistics 获取租户统计信息
func (s *SimpleTenantService) GetTenantStatistics(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// 统计总租户数
	totalCount, err := s.client.Tenant.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["total"] = totalCount

	// 统计根租户数
	rootCount, err := s.client.Tenant.Query().
		Where(tenant.ParentIDIsNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["root"] = rootCount

	// 统计子租户数
	subCount, err := s.client.Tenant.Query().
		Where(tenant.ParentIDNotNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["sub"] = subCount

	return stats, nil
}

// CanCreateSubTenant 检查是否可以创建子租户
func (s *SimpleTenantService) CanCreateSubTenant(ctx context.Context, tenantID int64) (bool, error) {
	t, err := s.client.Tenant.Get(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 只有集团型租户才能创建子租户
	return t.Type == tenant.TypeGROUP, nil
}
