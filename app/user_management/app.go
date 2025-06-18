package usermanagement

import (
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	v1 "github.com/yc-alpha/admin/api/user_management/v1"
	"github.com/yc-alpha/admin/app/user_management/internal/data"
	"github.com/yc-alpha/admin/app/user_management/internal/service"
)

func RegisteApplication(http *http.Server, grpc *grpc.Server) {
	basicData := data.NewData()
	basicData.Migrate()

	userService := service.NewUserService(basicData.Client)

	// Register HTTP services
	v1.RegisterUserServiceHTTPServer(http, userService)

	// Register gRPC services
	v1.RegisterUserServiceServer(grpc, userService)
}
