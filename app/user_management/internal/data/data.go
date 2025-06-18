package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	_ "github.com/lib/pq"
	"github.com/yc-alpha/admin/app/user_management/internal/data/ent"
	_ "github.com/yc-alpha/admin/app/user_management/internal/data/ent/runtime"
	"github.com/yc-alpha/config"
	"github.com/yc-alpha/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Data struct {
	Client *ent.Client
}

func NewData() *Data {
	return &Data{
		Client: NewDBClient(),
	}
}

func NewDBClient() *ent.Client {
	host := config.GetString("data.database.host", "")
	port := config.GetInt("data.database.port", 5432)
	username := config.GetString("data.database.username", "")
	password := config.GetString("data.database.password", "")
	db := config.GetString("data.database.db", "")
	url := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", host, port, username, db, password)
	client, err := ent.Open("postgres", url)
	if err != nil {
		logger.Fatalf("failed opening connection to postgres: %v", err)
	}
	// defer client.Close()
	return client
}

func (d *Data) Migrate() {
	// Run the auto migration tool.
	if err := d.Client.Schema.Create(context.Background()); err != nil {
		logger.Fatalf("failed creating schema resources: %v", err)
	}
	return
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
