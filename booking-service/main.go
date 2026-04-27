package main

import (
	"booking/booking-service/handler"
	"booking/booking-service/repository"
	"booking/booking-service/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// 1. MongoDB connection setup
	credential := options.Credential{
		Username: "booking_service",
		Password: "booking_service",
	}
	opts := options.Client().ApplyURI("mongodb://localhost:27028").SetAuth(credential)
	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal("Failed to disconnect from MongoDB:", err)
		}
	}()

	// 2. Initialize repository, service, and handler
	bookingsCollection := client.Database("booking").Collection("bookings")
	bookingRepo := repository.NewBookingRepository(bookingsCollection)
	bookingService := service.NewBookingService(bookingRepo)
	bookingHandler := handler.NewBookingHandler(bookingService)

	// 3. Start the HTTP server
	router := bookingHandler.Bookingrouter()
	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Welcome to Booking Service")
	})
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "healthy"})
	})

	srv := &http.Server{
		Addr:    ":8183",
		Handler: router,
	}

	go func() {
		log.Println("Server starting on :8183...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 4. Graceful shutdown on interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 5. Attempt graceful shutdown with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting gracefully")
}
