// @title           SJEK API
// @version         1.0.0.1
// @description     API server untuk SJEK

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// @default Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

package main

import (
	"log"
	_ "sjek/docs" // Import swagger docs
	"sjek/internal/database"
	"sjek/internal/routes"
)

func main() {
	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	router := routes.SetupRouter(db)

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
