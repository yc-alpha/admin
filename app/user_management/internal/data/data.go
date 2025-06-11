package data

import (
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/yc-alpha/config"
	"github.com/yc-alpha/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Data struct {
}

func NewData() *Data {
	return &Data{}
}

func NewRegistrar() registry.Registrar {
	host := config.GetString("registry.etcd.host", "")
	port := config.GetString("registry.etcd.port", "")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{host + ":" + port},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Fatal(err)
	}
	r := etcd.New(cli)
	return r
}
