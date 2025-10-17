package usermanagement

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	v1 "github.com/yc-alpha/admin/api/admin/v1"
	"github.com/yc-alpha/admin/app/admin/internal/config"
	"github.com/yc-alpha/admin/app/admin/internal/data"
	"github.com/yc-alpha/admin/app/admin/internal/service"
	"github.com/yc-alpha/logger"
)

func RegisteApplication(http *http.Server, grpc *grpc.Server) {
	basicData := data.NewData()
	basicData.InitDatabase(context.Background())

	// 初始化系统数据
	initService := service.NewInitService(basicData.Client)
	initConfig := config.LoadInitConfig()
	if err := initService.InitializeSystemWithConfig(context.Background(), initConfig); err != nil {
		logger.Fatalf("系统初始化失败: %v", err)
	}

	userService := service.NewUserService(basicData.Client)
	tenantHandler := service.NewTenantHTTPHandler(basicData.Client)

	// Register HTTP services
	v1.RegisterUserServiceHTTPServer(http, userService)
	http.HandleFunc("/v1/users/export", userService.ExportUser)

	// Register tenant HTTP handlers
	http.HandleFunc("/v1/tenants", tenantHandler.CreateTenant)
	http.HandleFunc("/v1/tenants/root", tenantHandler.ListRootTenants)
	http.HandleFunc("/v1/tenants/children", tenantHandler.ListSubTenants)
	http.HandleFunc("/v1/tenants/statistics", tenantHandler.GetTenantStatistics)
	http.HandleFunc("/v1/tenants/detail", tenantHandler.GetTenantByID)

	// Register gRPC services
	v1.RegisterUserServiceServer(grpc, userService)
}
