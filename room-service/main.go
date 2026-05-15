package main

import (
	"booking/room-service/handler"
	"booking/room-service/kafka"
	"booking/room-service/repository"
	"booking/room-service/service"
	"context"
	"fmt"
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

	//2. MongoDB connection setup
	credential := options.Credential{
		Username: "room_service",
		Password: "room_service",
	}
	opts := options.Client().ApplyURI("mongodb://localhost:27027").SetAuth(credential)
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
	roomsCollection := client.Database("booking").Collection("rooms")
	roomRepo := repository.NewRoomRepository(roomsCollection)
	roomService := service.NewRoomService(roomRepo)
	roomHandler := handler.NewRoomHandler(roomService)

	// 4. Start the HTTP server
	router := roomHandler.RoomRouter()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Room Service")
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	srv := &http.Server{
		Addr:    ":8182",
		Handler: router,
	}
	go func() {
		fmt.Println("Server starting on :8182...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// 5. Initialize and Start the Kafka Consumer
	log.Println("Kafka consumer starting ...")
	brokers := []string{kafka.BookingBrokerAddress}
	groupID := kafka.RoomReservationEventsGroupID
	topics := []string{kafka.BookingCreatedTopic}
	reservationProducer, err := kafka.NewReservationProducer(brokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	reservationConsumer, err := kafka.NewReservationConsumer(brokers, groupID, topics, roomService, reservationProducer)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	go reservationConsumer.Start(ctx)

	log.Println("Finished initialization, service is running...")

	// 6. Block the main thread waiting for the interrupt signal
	<-ctx.Done()
	log.Println("Shutdown signal received, exiting...")

	// 7. Attempt graceful shutdown with a timeout
	// This gives both the HTTP server and Kafka consumer 5 seconds to finish inflight tasks.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A. Shutdown HTTP server
	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server exiting gracefully")
	}
	// B. Shutdown Kafka consumer
	// Commits any pending offsets and closes connections to the broker.
	log.Println("Shutting down Kafka consumer...")
	reservationConsumer.Close()
	log.Println("Kafka consumer shut down successfully")

	// C. Shutdown Kafka producer
	// Flushes pending writes and closes connections to the broker.
	log.Println("Shutting down Kafka producer...")
	reservationProducer.Close()
	log.Println("Kafka producer shut down successfully")

	log.Println("Room Service exiting gracefully")
}
