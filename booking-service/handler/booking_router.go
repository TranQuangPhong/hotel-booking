package handler

import "github.com/gin-gonic/gin"

func (h *BookingHandler) Bookingrouter() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1/bookings")
	{
		v1.GET("/:id", h.GetBookingByID)
		v1.GET("/user/:userID", h.GetBookingsByUserID)
		v1.POST("/", h.CreateBooking)
		// TODO Cancel booking endpoint can be added here in the future
	}

	return r
}
