package server

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/yc-alpha/config"
	"time"
)

func NewGRPCServer() *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Network(config.GetString("server.grpc.network", "tcp")),
		grpc.Address(config.GetString("server.grpc.address", ":9100")),
		grpc.Timeout(time.Duration(config.GetInt64("server.grpc.timeout", 5000)) * time.Millisecond),
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	srv := grpc.NewServer(opts...)
	return srv
}
