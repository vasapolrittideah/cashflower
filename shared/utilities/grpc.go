package utilities

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

var defaultHeadersToForward = []string{
	"Authorization",
	"User-Agent",
	"X-Request-ID",
	"X-Forwarded-For",
	"X-Forwarded-Host",
	"X-Forwarded-Proto",
	"X-Real-IP",
}

// RegisterHealthServer registers the gRPC health check service.
func RegisterHealthServer(grpcServer *grpc.Server) {
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
}

// ForwardHTTPHeadersToGRPC extracts HTTP headers from the request and returns a context
// with gRPC metadata containing those headers. This allows the API gateway to forward
// headers to downstream gRPC services.
func ForwardHTTPHeadersToGRPC(ctx context.Context, r *http.Request, headersToForward []string) context.Context {
	md := metadata.New(nil)

	// Start with default headers and append any additional headers
	allHeaders := make([]string, len(defaultHeadersToForward))
	copy(allHeaders, defaultHeadersToForward)
	allHeaders = append(allHeaders, headersToForward...)

	// Remove duplicates by using a map
	seen := make(map[string]bool)
	for _, header := range allHeaders {
		if !seen[header] {
			seen[header] = true
			if values := r.Header.Values(header); len(values) > 0 {
				md.Set(header, values...)
			}
		}
	}

	return metadata.NewOutgoingContext(ctx, md)
}
