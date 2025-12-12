// @title           Campus Project Hub API
// @version         1.0
// @description     API for Campus Project Hub - Platform untuk berbagi project kampus
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://campus-project-hub.com/support
// @contact.email  support@campus-project-hub.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8000
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	_, err = database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Ensure upload directory exists
	if err := config.EnsureUploadDir(); err != nil {
		log.Printf("Warning: Failed to create upload directory: %v", err)
	}

	// Initialize router with all routes
	r := router.Setup(cfg)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// SIGINT (Ctrl+C) and SIGTERM (docker stop, kubernetes)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for shutdown (10 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	database.Close()

	log.Println("Server exited gracefully")
}
