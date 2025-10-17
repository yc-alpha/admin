package config

import (
	"github.com/yc-alpha/admin/app/admin/internal/service"
	"github.com/yc-alpha/config"
)

// LoadInitConfig 从配置文件加载初始化配置
func LoadInitConfig() *service.InitConfig {
	return &service.InitConfig{
		SystemTenantName:    config.GetString("system.init.tenant_name", "系统"),
		SystemTenantOwnerID: config.GetInt64("system.init.tenant_owner_id", 0),
		RootDeptName:        config.GetString("system.init.root_dept_name", "总公司"),
		AutoInit:            config.GetBool("system.init.auto_init", true),
	}
}
