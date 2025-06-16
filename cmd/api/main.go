package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"food-ordering-api/internal/app/dao"
	"food-ordering-api/internal/app/handler"
	"food-ordering-api/internal/app/logic"
	"food-ordering-api/internal/config"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Set package-level logger for handlers
	handler.SetLogger(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := initDB(*cfg)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize DAOs
	productDAO := dao.NewProductDAO(db)
	orderDAO := dao.NewOrderDAO(db)

	// Initialize logic layers
	productLogic := logic.NewProductLogic(productDAO)
	orderLogic := logic.NewOrderLogic(orderDAO, productDAO)

	// Initialize handlers
	productHandler := handler.NewProductHandler(productLogic)
	orderHandler := handler.NewOrderHandler(orderLogic)

	// Create a new Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Register API routes
	// Product routes
	router.GET("/api/products", productHandler.ListProducts)
	router.GET("/api/products/:id", productHandler.GetProduct)

	// Order routes
	router.POST("/api/orders", orderHandler.PlaceOrder)

	// Create the HTTP server
	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server with graceful shutdown
	go func() {
		logger.Info("starting server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the requests it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exiting")
}

// initDB initializes the database connection pool
func initDB(dbConfig config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbConfig.Database.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(dbConfig.Database.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.Database.MaxIdleConns)

	// Ping the database to verify the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
