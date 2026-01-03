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

	"backend/ent"
	"backend/internal/auth"
	"backend/internal/handler"
	"backend/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.Printf("[%s] %s %d %v - %s", v.Method, v.URI, v.Status, v.Latency, v.Error)
			} else {
				log.Printf("[%s] %s %d %v", v.Method, v.URI, v.Status, v.Latency)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", os.Getenv("FRONTEND_URL")},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "team_todo")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}

	// Run auto migration
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	log.Println("Database migration completed successfully")

	// Initialize services
	jwtService := auth.NewJWTService()
	emailService := service.NewEmailService()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(client, jwtService, emailService)
	orgHandler := handler.NewOrganizationHandler(client, emailService)
	projectHandler := handler.NewProjectHandler(client)
	contextHandler := handler.NewContextHandler(client)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// API routes
	api := e.Group("/api/v1")

	// Public routes
	api.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to Team Todo API",
		})
	})

	// Auth routes (public)
	authGroup := api.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)

	// Invite info (public - for showing invite details before login)
	api.GET("/invites/:token", orgHandler.GetInviteInfo)

	// Protected routes
	protected := api.Group("")
	protected.Use(auth.AuthMiddleware(jwtService))

	// User routes
	protected.GET("/me", authHandler.GetMe)
	protected.PATCH("/me", authHandler.UpdateMe)

	// Context routes
	protected.GET("/context", contextHandler.GetCurrentContext)
	protected.PUT("/context", contextHandler.UpdateContext)

	// Organization routes
	protected.POST("/organizations", orgHandler.CreateOrganization)
	protected.GET("/organizations", orgHandler.ListOrganizations)
	protected.GET("/organizations/:slug", orgHandler.GetOrganization)
	protected.POST("/organizations/:slug/invites", orgHandler.InviteMember)
	protected.POST("/invites/:token/accept", orgHandler.AcceptInvite)

	// Project routes
	protected.POST("/organizations/:slug/projects", projectHandler.CreateProject)
	protected.GET("/organizations/:slug/projects", projectHandler.ListProjects)
	protected.GET("/organizations/:slug/projects/:project_id", projectHandler.GetProject)
	protected.POST("/organizations/:slug/projects/:project_id/members", projectHandler.AddProjectMember)

	// Start server in a goroutine
	port := getEnv("PORT", "8080")
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	// Wait for interrupt signal (SIGINT or SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until signal is received

	log.Println("Shutting down server...")

	// Create a deadline for shutdown (10 seconds)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Close database connection
	if err := client.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
