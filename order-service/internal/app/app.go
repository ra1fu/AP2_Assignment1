package app

import (
	"database/sql"
	"fmt"
	nethttp "net/http"
	"time"

	"order-service/internal/repository"
	"order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// App represents the Order Service application.
type App struct {
	db     *sql.DB
	router *gin.Engine
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
	// Create HTTP client with 2-second timeout for Payment Service
	httpClient := &nethttp.Client{
		Timeout: 2 * time.Second,
	}

	orderRepo := repository.NewPostgresOrderRepository(db)
	paymentClient := repository.NewHTTPPaymentClient(paymentServiceURL, httpClient)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := http.NewOrderHandler(orderUC)

	// Setup routes
	http.SetupRoutes(router, orderHandler)

	return &App{
		db:     db,
		router: router,
	}, nil
}

// Run starts the Order Service server.
func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}

// Close closes the database connection.
func (a *App) Close() error {
	return a.db.Close()
}
