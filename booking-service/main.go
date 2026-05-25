package main

import (
	"booking/booking-service/handler"
	"booking/booking-service/kafka"
	"booking/booking-service/repository"
	"booking/booking-service/service"
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// 1. Create a root context that listens for OS shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 2. MongoDB connection setup
	credential := options.Credential{
		Username: "booking_service",
		Password: "booking_service",
	}
	opts := options.Client().
		ApplyURI("mongodb://mongo-booking-db:27028").
		SetAuth(credential).
		SetMaxPoolSize(5).
		SetMaxConnecting(5).
		SetMaxConnIdleTime(10 * time.Minute)
	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal("Failed to disconnect from MongoDB:", err)
		}
	}()

	// 3. Initialize repository, service, and handler
	bookingsCollection := client.Database("hotel-booking-system").Collection("bookings")
	bookingRepo := repository.NewBookingRepository(bookingsCollection)
	bookingService := service.NewBookingService(bookingRepo)
	bookingProducer, err := kafka.NewBookingProducer([]string{kafka.BookingBrokerAddress})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	bookingHandler := handler.NewBookingHandler(bookingService, bookingProducer)

	// 4. Start the HTTP server
	router := bookingHandler.Bookingrouter()
	router.GET("/bookings", func(c *gin.Context) {
		c.String(200, "Welcome to Booking Service")
	})
	router.GET("/bookings/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
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

	// 5. Initialize and Start the Kafka Consumer
	log.Println("Kafka consumer starting ...")
	brokers := []string{kafka.BookingBrokerAddress}
	groupID := kafka.RoomReservationEventsGroupID
	topics := []string{kafka.RoomReservationEventsTopic}
	consumer, err := kafka.NewRoomReservationConsumer(brokers, groupID, topics, bookingService)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	// Pass the OS signal context. The consumer will run until 'ctx' is canceled.
	go consumer.Start(ctx)

	log.Println("Finished initialization, service is running...")

	// 6. Block the main thread waiting for the interrupt signal
	<-ctx.Done()
	log.Println("Shutdown signal received, exiting...")

	// 7. Attempt graceful shutdown with a timeout
	// This gives both the HTTP server and Kafka consumer 5 seconds to finish inflight tasks.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A. Shutdown HTTP Server
	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server exiting gracefully")
	}
	// B. Close Kafka Consumer
	// Commits any pending offsets and closes connections to the broker.
	log.Println("Shutting down Kafka consumer...")
	consumer.Close()
	log.Println("Kafka consumer shut down successfully")

	// C. Close Kafka Producer
	// Flushes pending writes and closes connections to the broker.
	log.Println("Shutting down Kafka producer...")
	bookingProducer.Close()
	log.Println("Kafka producer shut down successfully")

	log.Println("Booking Service exiting gracefully")
}
