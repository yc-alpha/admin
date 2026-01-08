package config

import (
	"github.com/yc-alpha/config"
)

// InitConfig 初始化配置
type InitConfig struct {
	AutoInit            bool // 是否自动初始化
	SystemTenantName    string
	SystemTenantOwnerID int64
	RootDeptName        string
	RootName            string
	RootPassword        string
	RootFullName        string
}

// LoadInitConfig 从配置文件加载初始化配置
func LoadInitConfig() *InitConfig {
	return &InitConfig{
		AutoInit:            config.GetBool("system.init.auto_init", true),
		SystemTenantName:    config.GetString("system.init.tenant_name", "系统"),
		SystemTenantOwnerID: config.GetInt64("system.init.tenant_owner_id", 0),
		RootDeptName:        config.GetString("system.init.root_dept_name", "总公司"),
		RootName:            config.GetString("system.init.root_name", "admin"),
		RootPassword:        config.GetString("system.init.root_password", "Admin@2026"),
		RootFullName:        config.GetString("system.init.root_full_name", "系统管理员"),
	}
}
