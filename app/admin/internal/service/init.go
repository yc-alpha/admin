package service

import (
	"context"
	"fmt"

	"github.com/yc-alpha/admin/app/admin/internal/config"
	"github.com/yc-alpha/admin/app/admin/internal/constant"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/department"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/tenant"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/user"
	"github.com/yc-alpha/logger"
)

// InitService 负责应用启动时的初始化工作
type InitService struct {
	client *ent.Client
}

// DefaultInitConfig 默认初始化配置
func DefaultInitConfig() *config.InitConfig {
	return &config.InitConfig{
		AutoInit:            true,
		SystemTenantName:    "系统",
		SystemTenantOwnerID: 0,
		RootDeptName:        "总公司",
		RootName:            "admin",
		RootPassword:        "Admin@2026",
		RootFullName:        "系统管理员",
	}
}

// NewInitService 创建初始化服务
func NewInitService(client *ent.Client) *InitService {
	return &InitService{
		client: client,
	}
}

// InitializeSystem 初始化系统数据
func (s *InitService) InitializeSystem(ctx context.Context) error {
	return s.InitializeSystemWithConfig(ctx, DefaultInitConfig())
}

// InitializeSystemWithConfig 使用配置初始化系统数据
func (s *InitService) InitializeSystemWithConfig(ctx context.Context, config *config.InitConfig) error {
	if !config.AutoInit {
		logger.Info("自动初始化已禁用，跳过系统初始化")
		return nil
	}

	logger.Info("开始初始化系统数据...")

	// 1. 检查并创建系统租户
	systemTenant, err := s.ensureSystemTenantWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("创建系统租户失败: %w", err)
	}
	logger.Infof("系统租户已就绪，ID: %d", systemTenant.ID)

	// 2. 检查并创建总公司部门
	systemDept, err := s.ensureRootDepartmentWithConfig(ctx, systemTenant.ID, config)
	if err != nil {
		return fmt.Errorf("创建总公司部门失败: %w", err)
	}
	logger.Info("总公司部门已就绪")

	// 3. 检查并创建ROOT用户
	_, err = s.ensureRootUserWithConfig(ctx, systemTenant.ID, systemDept.ID, config)
	if err != nil {
		return fmt.Errorf("创建ROOT用户失败: %w", err)
	}
	logger.Info("ROOT用户已就绪")

	logger.Info("系统初始化完成")
	return nil
}

// ensureSystemTenantWithConfig 确保系统租户存在（使用指定配置）
func (s *InitService) ensureSystemTenantWithConfig(ctx context.Context, config *config.InitConfig) (*ent.Tenant, error) {
	// 检查是否已存在系统租户
	existingTenant, err := s.client.Tenant.Query().
		Where(tenant.Name(config.SystemTenantName)).
		First(ctx)

	if err == nil {
		logger.Info("系统租户已存在")
		return existingTenant, nil
	}

	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("查询系统租户失败: %w", err)
	}

	// 创建系统租户
	systemTenant, err := s.client.Tenant.Create().
		SetID(constant.ROOT_TENANT_ID).
		SetName(config.SystemTenantName).
		SetOwnerID(config.SystemTenantOwnerID).
		SetType(tenant.TypeROOT). // 设置为系统租户
		SetStatus(tenant.StatusACTIVE).
		SetLevel(0). // 根租户层级为0
		SetAttributes(map[string]any{
			"description": "系统默认租户",
			"type":        "system",
			"is_system":   true,
		}).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("创建系统租户失败: %w", err)
	}

	logger.Info("已创建系统租户（集团型）")
	return systemTenant, nil
}

// ensureRootDepartmentWithConfig 确保总公司部门存在（使用指定配置）
func (s *InitService) ensureRootDepartmentWithConfig(ctx context.Context, tenantID int64, config *config.InitConfig) (*ent.Department, error) {
	// 检查是否已存在总公司部门
	existingDept, err := s.client.Department.Query().
		Where(
			department.TenantID(tenantID),
			department.Name(config.RootDeptName),
		).
		First(ctx)

	if err == nil {
		logger.Info("总公司部门已存在")
		return existingDept, nil
	}

	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("查询总公司部门失败: %w", err)
	}

	// 创建总公司部门
	// 对于根部门，parent_id 设为 0，path 设为部门ID
	rootDept, err := s.client.Department.Create().
		SetTenantID(tenantID).
		SetParentID(0). // 根部门的父级ID为0
		SetName(config.RootDeptName).
		SetPath("0"). // 初始path设为"0"，创建后会更新为实际ID
		SetAttributes(map[string]any{
			"description": "系统默认根部门",
			"level":       0,
		}).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("创建总公司部门失败: %w", err)
	}

	// 更新部门的path为自身的ID（ltree格式）
	_, err = s.client.Department.UpdateOneID(rootDept.ID).
		SetPath(fmt.Sprintf("%d", rootDept.ID)).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("更新部门路径失败: %w", err)
	}

	logger.Info("已创建总公司部门")
	return rootDept, nil
}

func (s *InitService) ensureRootUserWithConfig(ctx context.Context, tenantID, deptID int64, config *config.InitConfig) (*ent.User, error) {
	// 检查是否已存在ROOT用户
	existingUser, err := s.client.User.Query().
		Where(user.Username(config.RootName)).First(ctx)
	if err == nil {
		logger.Info("ROOT用户已存在")
		return existingUser, nil
	}

	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("查询ROOT用户失败: %w", err)
	}

	// 创建ROOT用户
	// 开始事务
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	// 确保在函数结束时尝试回滚（若已提交，Rollback 会返回 sql.ErrTxDone）
	defer func() {
		_ = tx.Rollback() // 忽略错误处理示例（可按需记录）
	}()

	user, err := s.client.User.Create().
		SetUsername(config.RootName).
		SetPassword(config.RootPassword).
		SetFullName(config.RootFullName).
		SetStatus(user.StatusACTIVE).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("创建ROOT用户失败: %w", err)
	}

	_, err = tx.UserTenant.Create().
		SetUserID(user.ID).
		SetTenantID(tenantID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建ROOT用户租户关联失败: %w", err)
	}

	_, err = tx.UserDepartment.Create().
		SetUserID(user.ID).
		SetTenantID(tenantID).
		SetDepartmentID(deptID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建ROOT用户部门关联失败: %w", err)
	}

	// 显式提交
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("已创建ROOT用户")
	return user, nil
}

// CheckSystemStatus 检查系统状态
func (s *InitService) CheckSystemStatus(ctx context.Context) (bool, error) {
	// 检查是否有租户
	tenantCount, err := s.client.Tenant.Query().Count(ctx)
	if err != nil {
		return false, fmt.Errorf("检查租户数量失败: %w", err)
	}

	// 检查是否有部门
	deptCount, err := s.client.Department.Query().Count(ctx)
	if err != nil {
		return false, fmt.Errorf("检查部门数量失败: %w", err)
	}

	logger.Infof("系统状态检查: 租户数量=%d, 部门数量=%d", tenantCount, deptCount)

	// 如果租户和部门都为空，则需要初始化
	return tenantCount == 0 && deptCount == 0, nil
}
