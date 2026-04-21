package main

import (
	"log"
	"os"

	"order-service/internal/app"
)

func main() {
	// Configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "order_db")
	port := getEnv("PORT", ":8080")
	grpcPort := getEnv("GRPC_PORT", ":50051")
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "localhost:8081") // Note: gRPC usually doesn't use http://

	// Create and initialize the app
	orderApp, err := app.NewApp(dbHost, dbPort, dbUser, dbPassword, dbName, paymentServiceURL)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	defer orderApp.Close()

	log.Printf("Order Service REST API starting on %s", port)
	log.Printf("Order Service gRPC Server starting on %s", grpcPort)
	if err := orderApp.Run(port, grpcPort); err != nil {
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
