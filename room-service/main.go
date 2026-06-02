package main

import (
	"booking/room-service/handler"
	"booking/room-service/kafka"
	"booking/room-service/pkg/logger"
	"booking/room-service/repository"
	"booking/room-service/service"
	"context"
	"log/slog"
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
	// 1. Initialize structured logger FIRST
	log := logger.NewLogger()
	slog.SetDefault(log)

	// 2. Create a root context that listens for OS shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 3. MongoDB connection setup
	credential := options.Credential{
		Username: "room_service",
		Password: "room_service",
	}
	opts := options.Client().
		ApplyURI("mongodb://mongo-room-db:27027/?replicaSet=rs-room&readPreference=secondaryPreferred").
		SetAuth(credential).
		SetMaxPoolSize(5).
		SetMaxConnecting(5).
		SetMaxConnIdleTime(10 * time.Minute)
	client, err := mongo.Connect(opts)
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			slog.Error("Failed to disconnect from MongoDB", "error", err.Error())
		}
	}()

	// 4. Initialize repository, service, and handler
	roomsCollection := client.Database("hotel-booking-system").Collection("rooms")
	roomRepo := repository.NewRoomRepository(roomsCollection)

	inventoryCollection := client.Database("hotel-booking-system").Collection("inventory")
	inventoryRepo := repository.NewInventoryRepository(inventoryCollection)

	roomService := service.NewRoomService(client, roomRepo, inventoryRepo)
	roomHandler := handler.NewRoomHandler(roomService)

	// 5. Start the HTTP server
	router := roomHandler.RoomRouter()
	router.GET("/rooms", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Room Service")
	})
	router.GET("/rooms/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	srv := &http.Server{
		Addr:    ":8182",
		Handler: router,
	}
	go func() {
		slog.Info("Server starting", "addr", ":8182")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Listen error", "error", err.Error())
			os.Exit(1)
		}
	}()

	// 6. Initialize and Start the Kafka Consumer
	slog.Info("Kafka consumer starting")
	brokers := []string{kafka.BookingBrokerAddress}
	groupID := kafka.RoomReservationEventsGroupID
	topics := []string{kafka.BookingCreatedTopic}
	reservationProducer, err := kafka.NewReservationProducer(brokers)
	if err != nil {
		slog.Error("Failed to create Kafka producer", "error", err.Error())
		os.Exit(1)
	}
	reservationConsumer, err := kafka.NewReservationConsumer(brokers, groupID, topics, roomService, reservationProducer)
	if err != nil {
		slog.Error("Failed to create Kafka consumer", "error", err.Error())
		os.Exit(1)
	}
	go reservationConsumer.Start(ctx)

	slog.Info("Finished initialization, service is running")

	// 7. Block the main thread waiting for the interrupt signal
	<-ctx.Done()
	slog.Info("Shutdown signal received, exiting")

	// 8. Attempt graceful shutdown with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A. Shutdown HTTP server
	slog.Info("Shutting down HTTP server")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server forced to shutdown", "error", err.Error())
	} else {
		slog.Info("HTTP server exiting gracefully")
	}

	// B. Shutdown Kafka consumer
	slog.Info("Shutting down Kafka consumer")
	reservationConsumer.Close()
	slog.Info("Kafka consumer shut down successfully")

	// C. Shutdown Kafka producer
	slog.Info("Shutting down Kafka producer")
	reservationProducer.Close()
	slog.Info("Kafka producer shut down successfully")

	slog.Info("Room Service exiting gracefully")
}
