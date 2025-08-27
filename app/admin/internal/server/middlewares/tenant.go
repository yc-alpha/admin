package interceptors

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// 这是一个“示例”拦截器：
// 1) 从 metadata 中取 x-tenant-id
// 2) 开启 sql.Tx（与 ent client 共享同一 *sql.DB）
// 3) 在 tx 上执行 SET LOCAL app.current_tenant = $1
// 4) 将 tx 放入 ctx 中传给后续 handler（你的 repo 层应优先使用该 tx）
// 5) handler 返回成功则 Commit，否则 Rollback
//
// 注意：在实际系统中，你通常会把 ent.Client 传入并用 client.Tx(ctx) 开启 ent.Tx。
// 这里使用原生 *sql.DB 的 tx 来保证兼容性，并在 repo 层选择使用 tx.ExecContext / tx.QueryContext 或者用 ent 的 Tx 绑定。

type txKey struct{}

func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

// Unary interceptor
func RLSUnaryInterceptor(db *sql.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		vals := md.Get("x-tenant-id")
		if len(vals) == 0 {
			return nil, errors.New("missing x-tenant-id metadata")
		}
		tid, err := strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return nil, errors.New("invalid x-tenant-id")
		}

		tx, err := db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return nil, err
		}

		if _, err := tx.ExecContext(ctx, "SET LOCAL app.current_tenant = $1", tid); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		ctx = context.WithValue(ctx, txKey{}, tx)

		resp, err := handler(ctx, req)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return resp, nil
	}
}
