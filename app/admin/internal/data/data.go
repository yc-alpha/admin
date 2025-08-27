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
	if err := d.Client.Schema.Create(ctx); err != nil {
		logger.Fatalf("failed creating schema resources: %v", err)
	}
	// 启用 RLS
	err := d.InitRLS(ctx)
	if err != nil {
		logger.Fatalf("failed initializing RLS: %v", err)
	}
}

// InitRLS initializes Row-Level Security for tenant-scoped tables.
func (d *Data) InitRLS(ctx context.Context) error {
	// createTenantFunc creates a helper function `app_current_tenant()` to fetch current tenant_id from session.
	createFunc := `
		CREATE OR REPLACE FUNCTION app_current_tenant() RETURNS BIGINT AS $$
		BEGIN
			RETURN current_setting('app.current_tenant')::BIGINT;
		EXCEPTION WHEN others THEN
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql STABLE;
	`
	if _, err := d.DB.ExecContext(ctx, createFunc); err != nil {
		return fmt.Errorf("create app_current_tenant func: %v", err)
	}
	// list of tenant-scoped tables that must be protected by RLS
	tables := []string{"user_tenants", "departments", "user_departments"}
	for _, table := range tables {
		// enable rls
		if _, err := d.DB.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY;", table)); err != nil {
			return fmt.Errorf("enable rls on %s: %w", table, err)
		}
		// create policy
		// WITH CHECK ensures inserted/updated rows have tenant_id = current tenant
		policyName := table + "_rls"
		if _, err := d.DB.Exec(fmt.Sprintf(
			"CREATE OR REPLACE POLICY %s ON %s USING (tenant_id = app_current_tenant()) WITH CHECK (tenant_id = app_current_tenant());",
			policyName, table)); err != nil {
			return fmt.Errorf("create policy on %s: %w", table, err)
		}

	}
	return nil
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
