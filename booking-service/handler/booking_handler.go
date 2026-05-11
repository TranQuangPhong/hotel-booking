package handler

import (
	"booking/booking-service/kafka"
	"booking/booking-service/model"
	"booking/booking-service/service"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	service  *service.BookingService
	producer *kafka.BookingProducer
}

func NewBookingHandler(s *service.BookingService, p *kafka.BookingProducer) *BookingHandler {
	return &BookingHandler{service: s, producer: p}
}

func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	bookingID := c.Param("id")
	booking, err := h.service.GetBookingByID(c.Request.Context(), bookingID)

	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to fetch booking: %w", err).Error()})
		return
	}

	c.JSON(200, booking)
}

func (h *BookingHandler) GetBookingsByUserID(c *gin.Context) {
	userID := c.Param("userID")
	bookings, err := h.service.GetBookingsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to fetch bookings: %w", err).Error()})
	}
	c.JSON(200, bookings)
}

// Quickly publish a message to Kafka without waiting for DB confirmation.
// The consumer will handle the actual business logic and DB save.
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var booking *model.Booking //TODO: consider using a separate DTO for create requests
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(400, gin.H{"error": fmt.Errorf("invalid request body: %w", err).Error()})
		return
	}

	//Publish kafka message here instead of calling service method directly
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	err := h.producer.PublishCreateBookingMessage(ctx, booking)

	if err != nil {
		log.Printf("Failed to publish to Kafka: %v", err)
		c.JSON(500, gin.H{"error": fmt.Errorf("Service temporarily unavailable: failed to create booking: %w", err).Error()})
		return
	}
	c.JSON(201, gin.H{"message": "Booking created"})
}
