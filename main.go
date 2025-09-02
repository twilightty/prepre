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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"jinzmedia-atmt/auth"
	"jinzmedia-atmt/config"
	"jinzmedia-atmt/database"
	"jinzmedia-atmt/handlers"
	"jinzmedia-atmt/services"
)

func main() {
	// Load configuration
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := config.Get()

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()

	// Initialize services
	paymentService := services.NewPaymentService()
	authService := auth.NewAuthService()

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers()
	paymentHandlers := handlers.NewPaymentHandler(paymentService)
	downloadHandlers := handlers.NewDownloadHandlers()
	webhookHandlers := handlers.NewWebhookHandler(paymentService)
	adminHandlers := handlers.NewAdminHandlers()

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Timeout middleware
	r.Use(middleware.Timeout(cfg.Server.ReadTimeout))

	// Rate limiting (basic implementation)
	if cfg.RateLimit.RequestsPerMinute > 0 {
		r.Use(middleware.Throttle(cfg.RateLimit.RequestsPerMinute))
	}

	// Public routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no authentication required)
		r.Post("/auth/register", authHandlers.Register)
		r.Post("/auth/login", authHandlers.Login)
		r.Post("/auth/refresh", authHandlers.RefreshToken)

		// Protected routes (authentication required)
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware(authService))

			// User routes
			r.Get("/auth/profile", authHandlers.GetProfile)
			r.Post("/auth/logout", authHandlers.Logout)

			// Payment routes (authenticated users)
			r.Post("/payment/initiate", paymentHandlers.InitiatePayment)

			// Download routes (authenticated users)  
			r.Get("/download/{product_name}/{platform}", downloadHandlers.DownloadProduct)

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireAdmin())
				// Add admin-only routes here
			})

			// Super admin routes
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireSuper())
				// Add super admin-only routes here
			})
		})

		// Webhook routes (no authentication, but API key validation inside handler)
		r.Post("/hooks/sepay", webhookHandlers.HandleSepayWebhook)

		// Admin routes (no authentication required)
		r.Route("/admin", func(r chi.Router) {
			// Admin login
			r.Post("/login", adminHandlers.Login)

			// Dashboard
			r.Get("/dashboard/stats", adminHandlers.GetDashboardStats)

			// Analytics
			r.Get("/analytics/workflows/stats", adminHandlers.GetWorkflowStats)
			r.Get("/analytics/jobs/stats", adminHandlers.GetJobStats)
			r.Get("/analytics/costs/stats", adminHandlers.GetCostStats)

			// Jobs
			r.Get("/jobs", adminHandlers.GetJobs)
			r.Get("/jobs/{id}", adminHandlers.GetJob)

			// Workflows
			r.Get("/workflows", adminHandlers.GetWorkflows)
			r.Post("/workflows", adminHandlers.CreateWorkflow)
			r.Patch("/workflows/{id}", adminHandlers.UpdateWorkflow)
		})
	})

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("Hello from %s v%s!", cfg.App.Name, cfg.App.Version)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"message": "%s", "environment": "%s"}`, response, cfg.App.Environment)))
	})

	// Health check endpoint
	if cfg.HealthCheck.Enabled {
		r.Get(cfg.HealthCheck.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			// Check database connection
			dbHealthy := database.IsConnected()

			status := "healthy"
			statusCode := http.StatusOK
			if !dbHealthy {
				status = "unhealthy"
				statusCode = http.StatusServiceUnavailable
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write([]byte(fmt.Sprintf(`{
				"status": "%s", 
				"timestamp": "%s",
				"database": %t,
				"version": "%s"
			}`, status, time.Now().Format(time.RFC3339), dbHealthy, cfg.App.Version)))
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting %s v%s on %s", cfg.App.Name, cfg.App.Version, cfg.GetServerAddress())
		log.Printf("Database: %s", cfg.Database.Name)
		log.Printf("Environment: %s", cfg.App.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
