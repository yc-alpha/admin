// admin/common/middleware/authz.go
package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/yc-alpha/admin/common/authz"
)

// AuthzMiddleware Casbin授权中间件
func AuthzMiddleware(enforcer *casbin.Enforcer, subBuilder *authz.SubjectBuilder) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 1. 从context获取认证信息（假设已通过authn middleware）
			userID := GetUserIDFromContext(ctx)
			if userID == 0 {
				return nil, errors.Unauthorized("UNAUTHORIZED", "missing user authentication")
			}

			// 2. 获取租户ID（从header或context）
			tenantID := GetTenantIDFromContext(ctx)

			// 3. 构建Subject
			subject, err := subBuilder.BuildSubject(ctx, userID, tenantID)
			if err != nil {
				return nil, errors.InternalServer("AUTHZ_ERROR", err.Error())
			}

			// 4. 获取operation和method
			operation, method := extractOperationAndMethod(ctx)
			if operation == "" {
				// 无法获取operation，跳过授权（或根据策略拒绝）
				return handler(ctx, req)
			}

			// 5. Casbin Enforce
			domain := fmt.Sprintf("%d", tenantID)
			if tenantID == 0 {
				domain = "*" // 平台级API
			}

			ok, err := enforcer.Enforce(subject, domain, operation, method)
			if err != nil {
				return nil, errors.InternalServer("AUTHZ_ERROR", err.Error())
			}

			if !ok {
				return nil, errors.Forbidden("PERMISSION_DENIED",
					fmt.Sprintf("no permission to access %s", operation))
			}

			// 6. 如果有租户，设置PostgreSQL RLS上下文
			if tenantID > 0 {
				ctx = SetTenantContext(ctx, tenantID)
			}

			// 7. 将subject放入context供后续使用
			ctx = WithSubject(ctx, subject)

			return handler(ctx, req)
		}
	}
}

// extractOperationAndMethod 提取operation和method
func extractOperationAndMethod(ctx context.Context) (operation, method string) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		switch tr.Kind() {
		case transport.KindHTTP:
			if ht, ok := tr.(http.Transporter); ok {
				operation = ht.Operation()
				method = ht.Request().Method
			}
		case transport.KindGRPC:
			operation = tr.Operation()
			method = grpcMethodToHTTPMethod(operation)
		}
	}
	return
}

// grpcMethodToHTTPMethod 将gRPC方法映射为HTTP方法（用于统一授权）
func grpcMethodToHTTPMethod(operation string) string {
	lower := strings.ToLower(operation)
	switch {
	case strings.Contains(lower, "/create"), strings.Contains(lower, "/add"):
		return "POST"
	case strings.Contains(lower, "/update"), strings.Contains(lower, "/modify"):
		return "PUT"
	case strings.Contains(lower, "/delete"), strings.Contains(lower, "/remove"):
		return "DELETE"
	case strings.Contains(lower, "/get"), strings.Contains(lower, "/list"), strings.Contains(lower, "/query"):
		return "GET"
	default:
		return "POST" // 默认
	}
}

// Context helpers

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	tenantIDKey contextKey = "tenant_id"
	subjectKey  contextKey = "subject"
)

func GetUserIDFromContext(ctx context.Context) int64 {
	if v := ctx.Value(userIDKey); v != nil {
		return v.(int64)
	}
	return 0
}

func GetTenantIDFromContext(ctx context.Context) int64 {
	if v := ctx.Value(tenantIDKey); v != nil {
		return v.(int64)
	}
	return 0
}

func WithSubject(ctx context.Context, sub *authz.Subject) context.Context {
	return context.WithValue(ctx, subjectKey, sub)
}

func GetSubject(ctx context.Context) *authz.Subject {
	if v := ctx.Value(subjectKey); v != nil {
		return v.(*authz.Subject)
	}
	return nil
}

func SetTenantContext(ctx context.Context, tenantID int64) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}
