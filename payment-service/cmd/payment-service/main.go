package main

import (
	"log"
	"os"

	"payment-service/internal/app"
)

func main() {
	// Configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "payment_db")
	grpcPort := getEnv("GRPC_PORT", ":50052")

	// Create and initialize the app
	paymentApp, err := app.NewApp(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	defer paymentApp.Close()

	log.Printf("Payment Service gRPC Server starting on %s", grpcPort)
	if err := paymentApp.Run(grpcPort); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// getEnv gets environment variable or returns a default value.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
