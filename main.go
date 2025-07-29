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
	"io"
	"log"
	"os"
	_ "sjek/docs" // Import swagger docs
	"sjek/internal/database"
	"sjek/internal/routes"
	"gopkg.in/natefinch/lumberjack.v2"
)

func setupLogging() {
	// Create logs directory if not exists
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
		return
	}

	// Setup log rotation with timestamp in filename
	logRotator := &lumberjack.Logger{
		Filename:   "logs/sjek.log",
		MaxSize:    30,    // 30MB
		MaxBackups: 10,    // Keep 10 backup files
		MaxAge:     30,    // 30 days
		Compress:   true,  // Compress old files
		LocalTime:  true,  // Use local time for backup file timestamps
	}

	// Set log output to both file and console
	multiWriter := io.MultiWriter(os.Stdout, logRotator)
	log.SetOutput(multiWriter)
	
	// Set log format
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	log.Println("Logging initialized with 30MB rotation limit")
}

func main() {
	// Setup logging with rotation
	setupLogging()

	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	router := routes.SetupRouter(db)

	// Start server
	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
