package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

// LoggingInterceptor logs every incoming gRPC request with its method name and duration.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		log.Printf("[gRPC] --> Incoming Request: %s", info.FullMethod)

		m, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			log.Printf("[gRPC] <-- Response for %s failed with error: %v (took %v)", info.FullMethod, err, duration)
		} else {
			log.Printf("[gRPC] <-- Response for %s succeeded (took %v)", info.FullMethod, duration)
		}

		return m, err
	}
}
