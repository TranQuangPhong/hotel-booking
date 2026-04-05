package main

import (
	"booking/room-service/handler"
	"booking/room-service/repository"
	"booking/room-service/service"
	"context"
	"fmt"
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

	//1. MongoDB connection setup
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

	// 2. Initialize repository, service, and handler
	roomsCollection := client.Database("booking").Collection("rooms")
	roomRepo := repository.NewRoomRepository(roomsCollection)
	roomService := service.NewRoomService(roomRepo)
	roomHandler := handler.NewRoomHandler(roomService)

	// 3. Start the HTTP server
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

	// 4. Wait for interrupt signal to gracefully shutdown
	// quit channel listens for SIGINT (Ctrl+C) or SIGTERM (Kubernetes/Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 5. The Context is used to inform the server it has 5 seconds to finish
	// any existing requests before it is forced to close.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting gracefully")
}
