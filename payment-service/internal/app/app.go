package app

import (
	"database/sql"
	"fmt"

	"payment-service/internal/repository"
	"payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// App represents the Payment Service application.
type App struct {
	db     *sql.DB
	router *gin.Engine
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

	// Create router
	router := gin.Default()

	// Setup dependency injection (Composition Root)
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)
	paymentHandler := http.NewPaymentHandler(paymentUC)

	// Setup routes
	http.SetupRoutes(router, paymentHandler)

	return &App{
		db:     db,
		router: router,
	}, nil
}

// Run starts the Payment Service server.
func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}

// Close closes the database connection.
func (a *App) Close() error {
	return a.db.Close()
}
