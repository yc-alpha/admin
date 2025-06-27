package server

import (
	"time"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/yc-alpha/config"
)

func NewHTTPServer() *http.Server {
	var opts = []http.ServerOption{
		http.Network(config.GetString("server.http.network", "tcp")),
		http.Address(config.GetString("server.http.address", ":8010")),
		http.Timeout(time.Millisecond * time.Duration(config.GetInt64("server.http.timeout", 5000))),
		http.Middleware(
			recovery.Recovery(),
		),
	}
	srv := http.NewServer(opts...)
	return srv
}
