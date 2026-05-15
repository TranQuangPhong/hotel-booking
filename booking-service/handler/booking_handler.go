package handler

import (
	"booking/booking-service/event"
	"booking/booking-service/handler/dto"
	"booking/booking-service/kafka"
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

// Quickly save temp booking into DB.
// Publish msg to Room service to actually reserve room.
// TODO: to use outbox with CDC (next phase)
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req *dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Errorf("invalid request body: %w", err).Error()})
		return
	}

	// Convert DTO to domain model
	booking := req.ToModel()

	// Create temp booking immediately
	bookingID, err := h.service.CreateBooking(c.Request.Context(), booking)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to create booking: %w", err).Error()})
		return
	}
	msg := event.EventEnvelope[event.BookingCreatedMsg]{
		TraceID:   "0", //TODO: generate real trace ID for distributed tracing
		EventType: "booking_created",
		Producer:  "booking-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339), //Explicitly use UTC and RFC3339 for timestamps to ensure cross-service compatibility
		Data: event.BookingCreatedMsg{
			BookingID: bookingID,
			User: event.User{
				UserID:         booking.User.UserID,
				Name:           booking.User.Name,
				Email:          booking.User.Email,
				PhoneNumber:    booking.User.PhoneNumber,
				NumberOfGuests: booking.User.NumberOfGuests,
			},
			RoomID:       booking.Room.RoomID,
			CheckInDate:  booking.CheckInDate.Format(time.DateOnly),
			CheckOutDate: booking.CheckOutDate.Format(time.DateOnly),
			CreatedAt:    booking.CreatedAt.Format(time.RFC3339), // Use RFC3339 for timestamps to ensure cross-service compatibility
		},
	}

	// Publish kafka message
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	err = h.producer.PublishBookingCreated(ctx, msg)

	if err != nil {
		log.Printf("Failed to publish to Kafka: %v", err)
		c.JSON(500, gin.H{"error": fmt.Errorf("Service temporarily unavailable: failed to create booking: %w", err).Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Booking created", "id": bookingID})
}
