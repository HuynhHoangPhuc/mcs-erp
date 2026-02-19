package grpc

import (
	"context"
	"regexp"
	"strings"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var validSchema = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// reservedSchemas that must never be used as tenant identifiers.
var reservedSchemas = map[string]bool{
	"public": true, "pg_catalog": true, "information_schema": true,
	"pg_toast": true, "_template": true,
}

// TenantUnaryInterceptor extracts tenant from gRPC metadata and sets it in context.
func TenantUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		tenantIDs := md.Get("x-tenant-id")
		if len(tenantIDs) == 0 || tenantIDs[0] == "" {
			return nil, status.Error(codes.InvalidArgument, "missing x-tenant-id in metadata")
		}

		schema := strings.ReplaceAll(tenantIDs[0], "-", "_")
		if !validSchema.MatchString(schema) || reservedSchemas[schema] {
			return nil, status.Error(codes.InvalidArgument, "invalid tenant identifier")
		}

		ctx = tenant.WithTenant(ctx, schema)
		return handler(ctx, req)
	}
}
