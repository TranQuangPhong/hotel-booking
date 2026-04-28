package handler

import (
	"booking/booking-service/model"
	"booking/booking-service/service"
	"fmt"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	service *service.BookingService
}

func NewBookingHandler(s *service.BookingService) *BookingHandler {
	return &BookingHandler{service: s}
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

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var booking *model.Booking //TODO: consider using a separate DTO for create requests
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(400, gin.H{"error": fmt.Errorf("invalid request body: %w", err).Error()})
		return
	}
	err := h.service.CreateBooking(c.Request.Context(), booking)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to create booking: %w", err).Error()})
		return
	}
	c.JSON(201, gin.H{"message": "Booking created"})
}
