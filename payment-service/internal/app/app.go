package app

import (
	"database/sql"
	"fmt"
	"net"

	"payment-service/internal/repository"
	"payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"

	_ "github.com/lib/pq"
	paymentv1 "github.com/youruser/repo-b/payment/v1"
	mygrpc "google.golang.org/grpc"
)

// App represents the Payment Service application.
type App struct {
	db         *sql.DB
	grpcServer *mygrpc.Server
}

// NewApp creates and initializes a new Payment Service application.
func NewApp(dbHost, dbPort, dbUser, dbPassword, dbName string) (*App, error) {
	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Setup dependency injection (Composition Root)
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)
	paymentServer := grpc.NewPaymentServer(paymentUC)

	// Create and apply our interceptor (Bonus Point)
	interceptor := grpc.LoggingInterceptor()

	// Initialize gRPC Server
	gServer := mygrpc.NewServer(mygrpc.UnaryInterceptor(interceptor))
	paymentv1.RegisterPaymentServiceServer(gServer, paymentServer)

	return &App{
		db:         db,
		grpcServer: gServer,
	}, nil
}

// Run starts the Payment Service gRPC server.
func (a *App) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	fmt.Printf("Starting gRPC Payment Service on %s...\n", addr)
	return a.grpcServer.Serve(listener)
}

// Close closes the database connection and gracefully stops the gRPC Server.
func (a *App) Close() error {
	a.grpcServer.GracefulStop()
	return a.db.Close()
}
