package app

import (
	"database/sql"
	"fmt"
	"net"

	"order-service/internal/repository"
	transGrpc "order-service/internal/transport/grpc"
	"order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	orderv1 "github.com/youruser/repo-b/order/v1"
	mygrpc "google.golang.org/grpc"
)

// App represents the Order Service application.
type App struct {
	db         *sql.DB
	router     *gin.Engine
	grpcServer *mygrpc.Server
	payClient  *repository.GRPCPaymentClient
}

// NewApp creates and initializes a new Order Service application.
func NewApp(
	dbHost, dbPort, dbUser, dbPassword, dbName string,
	paymentServiceURL string,
) (*App, error) {
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

	// Create router
	router := gin.Default()

	// Setup dependency injection (Composition Root)
	// Create gRPC client for Payment Service
	paymentClient, err := repository.NewGRPCPaymentClient(paymentServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init grpc payment client: %w", err)
	}

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := http.NewOrderHandler(orderUC)

	// Setup routes
	http.SetupRoutes(router, orderHandler)

	// Setup gRPC Server for Streaming Order Updates
	gServer := mygrpc.NewServer()
	orderServer := transGrpc.NewOrderServer(orderUC)
	orderv1.RegisterOrderServiceServer(gServer, orderServer)

	return &App{
		db:         db,
		router:     router,
		grpcServer: gServer,
		payClient:  paymentClient,
	}, nil
}

// Run starts the Order Service server.
func (a *App) Run(httpAddr, grpcAddr string) error {
	// Start gRPC securely in a goroutine
	go func() {
		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			fmt.Printf("failed to listen for grpc on %s: %v\n", grpcAddr, err)
			return
		}
		fmt.Printf("Starting gRPC Order Service Streaming Server on %s...\n", grpcAddr)
		if err := a.grpcServer.Serve(listener); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	return a.router.Run(httpAddr)
}

// Close closes the database connection.
func (a *App) Close() error {
	a.grpcServer.GracefulStop()
	if a.payClient != nil {
		_ = a.payClient.Close()
	}
	return a.db.Close()
}
