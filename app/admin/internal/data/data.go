package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	_ "github.com/lib/pq"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent"
	"github.com/yc-alpha/admin/app/admin/internal/data/ent/migrate"
	_ "github.com/yc-alpha/admin/app/admin/internal/data/ent/runtime"
	"github.com/yc-alpha/config"
	"github.com/yc-alpha/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Data struct {
	Client *ent.Client
	DB     *sql.DB
}

func NewData() *Data {
	client, db := NewDBClient()
	return &Data{
		Client: client,
		DB:     db,
	}
}

func NewDBClient() (*ent.Client, *sql.DB) {
	host := config.GetString("data.database.host", "")
	port := config.GetInt("data.database.port", 5432)
	username := config.GetString("data.database.username", "")
	password := config.GetString("data.database.password", "")
	dbName := config.GetString("data.database.db", "")
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", host, port, username, dbName, password)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatalf("failed opening connection to postgres: %v", err)
	}
	// 在 DB 上确保 ltree 扩展存在（运行时创建）
	if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS ltree"); err != nil {
		// log and continue or return error based on your policy
		logger.Fatalf("warn: failed to create ltree extension: %v", err)
	}
	// 用 ent 的 SQL driver 封装
	drv := entsql.OpenDB("postgres", db)
	// 创建 ent client
	client := ent.NewClient(ent.Driver(drv))
	// defer client.Close()
	return client, db
}

func (d *Data) Migrate(ctx context.Context) {
	// Run the auto migration tool.
	if err := d.Client.Schema.Create(ctx,
		migrate.WithDropIndex(false),
		migrate.WithDropColumn(false),
	); err != nil {
		logger.Fatalf("failed creating schema resources: %v", err)
	}
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
