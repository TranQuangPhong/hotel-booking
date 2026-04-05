package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *RoomHandler) RoomRouter() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1/rooms")
	{
		// Room routes
		v1.GET("/", h.GetAllRooms)         // Get all rooms
		v1.GET("/:id", h.GetRoomByID)      // Get room by ID
		v1.POST("/", h.CreateRoom)         // Create a new room
		v1.PUT("/:id", h.UpdateRoomStatus) // Update room
		// v1.DELETE("/:id", h.DeleteRoom) // Delete room
	}

	return r
}
