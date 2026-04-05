package handler

import (
	"booking/room-service/model"
	"booking/room-service/service"
	"fmt"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	service *service.RoomService
}

func NewRoomHandler(s *service.RoomService) *RoomHandler {
	return &RoomHandler{service: s}
}

// Get all rooms
func (h *RoomHandler) GetAllRooms(c *gin.Context) {
	rooms, err := h.service.GetAllRooms(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to fetch rooms: %w", err).Error()})
		return
	}

	c.JSON(200, rooms)
}

// Get room by ID
func (h *RoomHandler) GetRoomByID(c *gin.Context) {
	roomID := c.Param("id")
	room, err := h.service.GetRoomByID(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(404, gin.H{"error": fmt.Errorf("failed to fetch room: %w", err).Error()})
		return
	}
	c.JSON(200, room)
}

// Create new room
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var room *model.Room //TODO: consider using a separate DTO for create requests
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(400, gin.H{"error": fmt.Errorf("invalid request body: %w", err).Error()})
		return
	}
	err := h.service.CreateRoom(c.Request.Context(), room)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to create room: %w", err).Error()})
		return
	}
	c.JSON(201, gin.H{"message": "Room created"})
}

// Update room status
func (h *RoomHandler) UpdateRoomStatus(c *gin.Context) {

	roomID := c.Param("id")
	// Define a small DTO for the request body
	var input struct {
		Status model.RoomStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": fmt.Errorf("invalid request body: %w", err).Error()})
		return
	}

	err := h.service.UpdateRoomStatus(c.Request.Context(), roomID, input.Status)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Errorf("failed to update room status: %w", err).Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Room status updated"})
}
