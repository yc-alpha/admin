package service

import (
	"context"
	"strconv"
	"time"

	v1 "github.com/yc-alpha/admin/api/admin/v1"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent"
	"github.com/yc-alpha/logger"
)

// TenantServiceImpl 租户服务实现
type TenantServiceImpl struct {
	v1.UnimplementedTenantServiceServer
	tenantService *TenantService
}

// NewTenantServiceImpl 创建租户服务实现
func NewTenantServiceImpl(client *ent.Client) *TenantServiceImpl {
	return &TenantServiceImpl{
		tenantService: NewTenantService(client),
	}
}

// CreateTenant 创建租户
func (s *TenantServiceImpl) CreateTenant(ctx context.Context, req *v1.CreateTenantRequest) (*v1.CreateTenantResponse, error) {
	// 转换请求参数
	ownerID, err := strconv.ParseInt(req.OwnerId, 10, 64)
	if err != nil {
		return &v1.CreateTenantResponse{
			Result: false,
			Code:   400,
			Msg:    "无效的拥有者ID",
		}, nil
	}

	var parentID *int64
	if req.ParentId != "" {
		pid, err := strconv.ParseInt(req.ParentId, 10, 64)
		if err != nil {
			return &v1.CreateTenantResponse{
				Result: false,
				Code:   400,
				Msg:    "无效的父租户ID",
			}, nil
		}
		parentID = &pid
	}

	var createdBy *int64
	if req.CreatedBy != "" {
		cb, err := strconv.ParseInt(req.CreatedBy, 10, 64)
		if err != nil {
			return &v1.CreateTenantResponse{
				Result: false,
				Code:   400,
				Msg:    "无效的创建者ID",
			}, nil
		}
		createdBy = &cb
	}

	// 转换属性
	attributes := make(map[string]any)
	for k, v := range req.Attributes {
		attributes[k] = v
	}

	// 创建租户请求
	createReq := &CreateTenantRequest{
		Name:       req.Name,
		OwnerID:    ownerID,
		Type:       TenantType(req.Type.String()),
		ParentID:   parentID,
		Status:     req.Status.String(),
		Attributes: attributes,
		CreatedBy:  createdBy,
	}

	// 处理过期时间
	if req.ExpiredAt != "" {
		expiredAt, err := time.Parse(time.RFC3339, req.ExpiredAt)
		if err == nil {
			expiredAtStr := expiredAt.Format(time.RFC3339)
			createReq.ExpiredAt = &expiredAtStr
		}
	}

	// 调用服务创建租户
	resp, err := s.tenantService.CreateTenant(ctx, createReq)
	if err != nil {
		logger.Errorf("创建租户失败: %v", err)
		return &v1.CreateTenantResponse{
			Result: false,
			Code:   500,
			Msg:    "创建租户失败",
		}, nil
	}

	if !resp.Success {
		return &v1.CreateTenantResponse{
			Result: false,
			Code:   400,
			Msg:    resp.Message,
		}, nil
	}

	// 转换响应
	tenantProto := s.convertTenantToProto(resp.Tenant)
	return &v1.CreateTenantResponse{
		Result: true,
		Code:   200,
		Msg:    "租户创建成功",
		Tenant: tenantProto,
	}, nil
}

// GetTenant 获取租户详情
func (s *TenantServiceImpl) GetTenant(ctx context.Context, req *v1.GetTenantRequest) (*v1.GetTenantResponse, error) {
	tenantID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return &v1.GetTenantResponse{
			Result: false,
			Code:   400,
			Msg:    "无效的租户ID",
		}, nil
	}

	tenant, err := s.tenantService.client.Tenant.Get(ctx, tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return &v1.GetTenantResponse{
				Result: false,
				Code:   404,
				Msg:    "租户不存在",
			}, nil
		}
		return &v1.GetTenantResponse{
			Result: false,
			Code:   500,
			Msg:    "查询租户失败",
		}, nil
	}

	tenantProto := s.convertTenantToProto(tenant)
	return &v1.GetTenantResponse{
		Result: true,
		Code:   200,
		Msg:    "查询成功",
		Tenant: tenantProto,
	}, nil
}

// GetTenantHierarchy 获取租户层级结构
func (s *TenantServiceImpl) GetTenantHierarchy(ctx context.Context, req *v1.GetTenantHierarchyRequest) (*v1.GetTenantHierarchyResponse, error) {
	tenantID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return &v1.GetTenantHierarchyResponse{
			Result: false,
			Code:   400,
			Msg:    "无效的租户ID",
		}, nil
	}

	tenant, err := s.tenantService.GetTenantHierarchy(ctx, tenantID)
	if err != nil {
		if ent.IsNotFound(err) {
			return &v1.GetTenantHierarchyResponse{
				Result: false,
				Code:   404,
				Msg:    "租户不存在",
			}, nil
		}
		return &v1.GetTenantHierarchyResponse{
			Result: false,
			Code:   500,
			Msg:    "查询租户层级失败",
		}, nil
	}

	tenantProto := s.convertTenantToProto(tenant)
	return &v1.GetTenantHierarchyResponse{
		Result: true,
		Code:   200,
		Msg:    "查询成功",
		Tenant: tenantProto,
	}, nil
}

// ListRootTenants 获取根租户列表
func (s *TenantServiceImpl) ListRootTenants(ctx context.Context, req *v1.ListRootTenantsRequest) (*v1.ListRootTenantsResponse, error) {
	tenants, err := s.tenantService.GetRootTenants(ctx)
	if err != nil {
		return &v1.ListRootTenantsResponse{
			Result: false,
			Code:   500,
			Msg:    "查询根租户失败",
		}, nil
	}

	var tenantProtos []*v1.Tenant
	for _, t := range tenants {
		tenantProtos = append(tenantProtos, s.convertTenantToProto(t))
	}

	return &v1.ListRootTenantsResponse{
		Result:  true,
		Code:    200,
		Msg:     "查询成功",
		Tenants: tenantProtos,
		Total:   int32(len(tenantProtos)),
	}, nil
}

// ListSubTenants 获取子租户列表
func (s *TenantServiceImpl) ListSubTenants(ctx context.Context, req *v1.ListSubTenantsRequest) (*v1.ListSubTenantsResponse, error) {
	parentID, err := strconv.ParseInt(req.ParentId, 10, 64)
	if err != nil {
		return &v1.ListSubTenantsResponse{
			Result: false,
			Code:   400,
			Msg:    "无效的父租户ID",
		}, nil
	}

	tenants, err := s.tenantService.GetSubTenants(ctx, parentID)
	if err != nil {
		return &v1.ListSubTenantsResponse{
			Result: false,
			Code:   500,
			Msg:    "查询子租户失败",
		}, nil
	}

	var tenantProtos []*v1.Tenant
	for _, t := range tenants {
		tenantProtos = append(tenantProtos, s.convertTenantToProto(t))
	}

	return &v1.ListSubTenantsResponse{
		Result:  true,
		Code:    200,
		Msg:     "查询成功",
		Tenants: tenantProtos,
		Total:   int32(len(tenantProtos)),
	}, nil
}

// ListGroupTenants 获取集团型租户列表
func (s *TenantServiceImpl) ListGroupTenants(ctx context.Context, req *v1.ListGroupTenantsRequest) (*v1.ListGroupTenantsResponse, error) {
	tenants, err := s.tenantService.GetGroupTenants(ctx)
	if err != nil {
		return &v1.ListGroupTenantsResponse{
			Result: false,
			Code:   500,
			Msg:    "查询集团型租户失败",
		}, nil
	}

	var tenantProtos []*v1.Tenant
	for _, t := range tenants {
		tenantProtos = append(tenantProtos, s.convertTenantToProto(t))
	}

	return &v1.ListGroupTenantsResponse{
		Result:  true,
		Code:    200,
		Msg:     "查询成功",
		Tenants: tenantProtos,
		Total:   int32(len(tenantProtos)),
	}, nil
}

// GetTenantStatistics 获取租户统计信息
func (s *TenantServiceImpl) GetTenantStatistics(ctx context.Context, req *v1.GetTenantStatisticsRequest) (*v1.GetTenantStatisticsResponse, error) {
	stats, err := s.tenantService.GetTenantStatistics(ctx)
	if err != nil {
		return &v1.GetTenantStatisticsResponse{
			Result: false,
			Code:   500,
			Msg:    "获取统计信息失败",
		}, nil
	}

	statistics := make(map[string]int32)
	for k, v := range stats {
		statistics[k] = int32(v)
	}

	return &v1.GetTenantStatisticsResponse{
		Result:     true,
		Code:       200,
		Msg:        "查询成功",
		Statistics: statistics,
	}, nil
}

// UpdateTenant 更新租户
func (s *TenantServiceImpl) UpdateTenant(ctx context.Context, req *v1.UpdateTenantRequest) (*v1.UpdateTenantResponse, error) {
	tenantID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return &v1.UpdateTenantResponse{
			Result: false,
			Code:   400,
			Msg:    "无效的租户ID",
		}, nil
	}

	// 构建更新器
	updater := s.tenantService.client.Tenant.UpdateOneID(tenantID)

	if req.Name != "" {
		updater.SetName(req.Name)
	}

	if req.OwnerId != "" {
		ownerID, err := strconv.ParseInt(req.OwnerId, 10, 64)
		if err != nil {
			return &v1.UpdateTenantResponse{
				Result: false,
				Code:   400,
				Msg:    "无效的拥有者ID",
			}, nil
		}
		updater.SetOwnerID(ownerID)
	}
	// TODO:此处报错，待修复
	// if req.Status != v1.TenantStatus_TENANT_STATUS_UNSPECIFIED {
	// 	updater.SetStatus(tenant.Status(req.Status.String()))
	// }

	if req.ExpiredAt != "" {
		expiredAt, err := time.Parse(time.RFC3339, req.ExpiredAt)
		if err == nil {
			updater.SetExpiredAt(expiredAt)
		}
	}

	if req.Attributes != nil {
		attributes := make(map[string]any)
		for k, v := range req.Attributes {
			attributes[k] = v
		}
		updater.SetAttributes(attributes)
	}

	if req.UpdatedBy != "" {
		updatedBy, err := strconv.ParseInt(req.UpdatedBy, 10, 64)
		if err == nil {
			updater.SetUpdatedBy(updatedBy)
		}
	}

	// 执行更新
	updatedTenant, err := updater.Save(ctx)
	if err != nil {
		return &v1.UpdateTenantResponse{
			Result: false,
			Code:   500,
			Msg:    "更新租户失败",
		}, nil
	}

	tenantProto := s.convertTenantToProto(updatedTenant)
	return &v1.UpdateTenantResponse{
		Result: true,
		Code:   200,
		Msg:    "更新成功",
		Tenant: tenantProto,
	}, nil
}

// DeleteTenant 删除租户
func (s *TenantServiceImpl) DeleteTenant(ctx context.Context, req *v1.DeleteTenantRequest) (*v1.DeleteTenantResponse, error) {
	tenantID, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return &v1.DeleteTenantResponse{
			Result: false,
			Code:   400,
			Msg:    "无效的租户ID",
		}, nil
	}

	// 检查是否有子租户
	subTenants, err := s.tenantService.GetSubTenants(ctx, tenantID)
	if err == nil && len(subTenants) > 0 {
		return &v1.DeleteTenantResponse{
			Result: false,
			Code:   400,
			Msg:    "该租户下还有子租户，无法删除",
		}, nil
	}

	// 执行软删除
	err = s.tenantService.client.Tenant.UpdateOneID(tenantID).
		SetDeletedAt(time.Now()).
		Exec(ctx)

	if err != nil {
		return &v1.DeleteTenantResponse{
			Result: false,
			Code:   500,
			Msg:    "删除租户失败",
		}, nil
	}

	return &v1.DeleteTenantResponse{
		Result: true,
		Code:   200,
		Msg:    "删除成功",
	}, nil
}

// convertTenantToProto 转换租户实体为Proto消息
func (s *TenantServiceImpl) convertTenantToProto(t *ent.Tenant) *v1.Tenant {
	tenantProto := &v1.Tenant{
		Id:        strconv.FormatInt(t.ID, 10),
		Name:      t.Name,
		OwnerId:   strconv.FormatInt(t.OwnerID, 10),
		Type:      v1.TenantType(v1.TenantType_value[t.Type.String()]),
		Level:     int32(t.Level),
		Status:    v1.TenantStatus(v1.TenantStatus_value[t.Status.String()]),
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		Deleted:   t.DeletedAt != nil,
	}

	if t.ParentID != nil {
		tenantProto.ParentId = strconv.FormatInt(*t.ParentID, 10)
	}

	if t.Path != nil {
		tenantProto.Path = *t.Path
	}

	if t.ExpiredAt != nil {
		tenantProto.ExpiredAt = t.ExpiredAt.Format(time.RFC3339)
	}

	if t.CreatedBy != nil {
		tenantProto.CreatedBy = strconv.FormatInt(*t.CreatedBy, 10)
	}

	if t.UpdatedBy != nil {
		tenantProto.UpdatedBy = strconv.FormatInt(*t.UpdatedBy, 10)
	}

	// 转换属性
	tenantProto.Attributes = make(map[string]string)
	for k, v := range t.Attributes {
		if str, ok := v.(string); ok {
			tenantProto.Attributes[k] = str
		}
	}

	// 转换子租户
	if t.Edges.Children != nil {
		for _, child := range t.Edges.Children {
			tenantProto.Children = append(tenantProto.Children, s.convertTenantToProto(child))
		}
	}

	// 转换父租户
	if t.Edges.Parent != nil {
		tenantProto.Parent = s.convertTenantToProto(t.Edges.Parent)
	}

	return tenantProto
}
