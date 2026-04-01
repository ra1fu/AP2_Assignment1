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
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081")

	// Create and initialize the app
	orderApp, err := app.NewApp(dbHost, dbPort, dbUser, dbPassword, dbName, paymentServiceURL)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	defer orderApp.Close()

	log.Printf("Order Service starting on %s", port)
	if err := orderApp.Run(port); err != nil {
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
