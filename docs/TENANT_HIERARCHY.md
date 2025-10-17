# 租户层级管理功能

## 概述

租户层级管理功能支持三种类型的租户：普通租户、集团型租户和子租户，实现了完整的租户层级结构管理。

## 租户类型

### 1. 普通租户 (NORMAL)
- **特点**: 独立的租户，没有父级关系
- **用途**: 适用于单一组织或小型企业
- **限制**: 不能创建子租户

### 2. 集团型租户 (GROUP)
- **特点**: 根级租户，可以创建子租户
- **用途**: 适用于大型集团企业
- **权限**: 可以管理下属的所有子租户

### 3. 子租户 (SUB)
- **特点**: 必须有父级租户，父级必须是集团型租户
- **用途**: 适用于集团下属的子公司或部门
- **限制**: 不能创建子租户，是层级结构的叶节点

## 数据库结构

### 租户表字段

```sql
CREATE TABLE tenants (
    id BIGINT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    owner_id BIGINT NOT NULL,
    type VARCHAR(10) NOT NULL,           -- NORMAL, GROUP, SUB
    parent_id BIGINT NULL,              -- 父租户ID（子租户专用）
    path LTREE NULL,                    -- 层级路径（ltree格式）
    level INTEGER DEFAULT 0,            -- 层级深度
    status VARCHAR(10) DEFAULT 'PENDING',
    expired_at TIMESTAMPTZ NULL,
    attributes JSONB DEFAULT '{}',
    created_by BIGINT NULL,
    updated_by BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);
```

### 约束条件

1. **类型约束**: 只有集团型租户才能有子租户
2. **层级约束**: 子租户必须指定父租户
3. **路径约束**: 使用 ltree 格式存储层级关系

## API 接口

### HTTP 接口

#### 创建租户
```http
POST /v1/tenants
Content-Type: application/json

{
    "name": "租户名称",
    "owner_id": 1,
    "type": "GROUP",
    "parent_id": null
}
```

#### 获取根租户列表
```http
GET /v1/tenants/root
```

#### 获取子租户列表
```http
GET /v1/tenants/children?parent_id=123
```

#### 获取租户统计信息
```http
GET /v1/tenants/statistics
```

#### 获取租户详情
```http
GET /v1/tenants/detail?id=123
```

### 响应格式

#### 成功响应
```json
{
    "success": true,
    "message": "操作成功",
    "tenant": {
        "id": "1234567890",
        "name": "租户名称",
        "owner_id": "1",
        "type": "GROUP",
        "parent_id": null,
        "path": "1234567890",
        "level": 0,
        "status": "ACTIVE",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

#### 错误响应
```json
{
    "success": false,
    "message": "错误信息"
}
```

## 使用示例

### 1. 创建集团型租户

```go
// 创建集团型租户
tenantService := NewSimpleTenantService(client)
groupTenant, err := tenantService.CreateTenant(ctx, "ABC集团", 1, "GROUP")
if err != nil {
    log.Fatal(err)
}
```

### 2. 创建子租户

```go
// 创建子租户
subTenant, err := tenantService.CreateTenant(ctx, "ABC子公司", 1, "SUB")
if err != nil {
    log.Fatal(err)
}
```

### 3. 查询租户层级

```go
// 获取根租户列表
rootTenants, err := tenantService.GetRootTenants(ctx)
if err != nil {
    log.Fatal(err)
}

// 获取子租户列表
subTenants, err := tenantService.GetSubTenants(ctx, parentID)
if err != nil {
    log.Fatal(err)
}
```

### 4. 获取统计信息

```go
// 获取租户统计
stats, err := tenantService.GetTenantStatistics(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("总租户数: %d\n", stats["total"])
fmt.Printf("根租户数: %d\n", stats["root"])
fmt.Printf("子租户数: %d\n", stats["sub"])
```

## 层级关系示例

```
系统租户 (GROUP, level=0)
├── ABC集团 (GROUP, level=0)
│   ├── ABC北京公司 (SUB, level=1)
│   ├── ABC上海公司 (SUB, level=1)
│   └── ABC深圳公司 (SUB, level=1)
├── XYZ集团 (GROUP, level=0)
│   ├── XYZ科技 (SUB, level=1)
│   └── XYZ贸易 (SUB, level=1)
└── 独立公司A (NORMAL, level=0)
```

## 权限控制

### 租户隔离
- 每个租户的数据完全隔离
- 子租户只能访问自己租户下的数据
- 集团型租户可以管理所有子租户

### 数据访问
- 使用 RLS (Row Level Security) 实现数据隔离
- 基于租户ID进行数据过滤
- 支持跨租户数据查询（需要特殊权限）

## 最佳实践

### 1. 租户设计
- 集团型租户作为顶层容器
- 子租户按业务或地域划分
- 避免过深的层级结构（建议不超过3层）

### 2. 性能优化
- 使用 ltree 索引优化层级查询
- 合理设置租户数量限制
- 定期清理无效租户数据

### 3. 安全考虑
- 严格控制租户创建权限
- 实现租户数据备份和恢复
- 监控租户资源使用情况

## 故障排除

### 常见问题

1. **租户创建失败**
   - 检查租户类型约束
   - 验证父租户是否存在
   - 确认权限设置

2. **层级查询错误**
   - 检查 ltree 路径格式
   - 验证租户关系完整性
   - 确认索引状态

3. **权限访问问题**
   - 检查租户ID设置
   - 验证RLS策略
   - 确认用户权限

### 调试方法

```go
// 检查租户层级
path, err := tenantService.GetTenantPath(ctx, tenantID)
if err != nil {
    log.Printf("获取租户路径失败: %v", err)
}

// 验证租户类型
canCreate, err := tenantService.CanCreateSubTenant(ctx, tenantID)
if err != nil {
    log.Printf("检查子租户创建权限失败: %v", err)
}
```

## 扩展功能

### 未来计划
- 租户数据迁移工具
- 租户资源监控面板
- 多租户数据同步
- 租户权限管理界面

### 自定义扩展
- 租户自定义属性
- 租户间数据共享
- 租户生命周期管理
- 租户计费统计
