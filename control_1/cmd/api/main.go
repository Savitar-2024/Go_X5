package main

import (
	"log/slog"
	"net/http"
	"os"

	_ "example.com/order-service/docs"
	"example.com/order-service/internal/handler"
	"example.com/order-service/internal/repository/memory"
	"example.com/order-service/internal/service"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Orders API
// @version 1.0
// @description HTTP API для управления заказами и их статусами.
// @host localhost:9090
// @BasePath /
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	repository := memory.NewOrderRepository()
	orderService := service.NewOrderService(repository)
	router := handler.NewRouter(orderService, logger)

	// Добавляем Swagger UI к существующему HTTP-обработчику.
	mux := http.NewServeMux()
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/", router)

	address := envOrDefault("HTTP_ADDR", ":8080")
	logger.Info("starting HTTP server", slog.String("address", address))

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
