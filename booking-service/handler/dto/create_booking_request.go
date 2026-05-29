package dto

import (
	"booking/booking-service/model"
	"encoding/json"
	"fmt"
	"time"
)

// CreateBookingRequest is the DTO for incoming create-booking requests.
// It only exposes fields that the client is allowed to provide.
// TotalAmount and NightlyRates are NOT accepted — room-service is the sole price authority for final pricing.
// Room.Price and Room.Currency are snapshot values from the client's search results (display only, not final).
type CreateBookingRequest struct {
	User         User     `json:"user" binding:"required"`
	Room         Room     `json:"room" binding:"required"`
	CheckInDate  DateOnly `json:"check_in_date" binding:"required"`
	CheckOutDate DateOnly `json:"check_out_date" binding:"required"`
}

type User struct {
	UserID         string `json:"user_id" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	PhoneNumber    string `json:"phone_number" binding:"required,e164"`
	NumberOfGuests int    `json:"number_of_guests" binding:"required,gt=0"`
}

type Room struct {
	RoomID     string  `json:"room_id" binding:"required"`
	RoomNumber string  `json:"room_number" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	Price      float64 `json:"price" binding:"required,gt=0"`
	Currency   string  `json:"currency" binding:"required"`
}

type DateOnly time.Time

// MarshalJSON serializes the date to "YYYY-MM-DD" JSON format.
func (d DateOnly) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%q", t.Format(time.DateOnly))), nil
}

// UnmarshalJSON deserializes a "YYYY-MM-DD" JSON string into time.Time.
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}
	*d = DateOnly(t)
	return nil
}

// ToModel converts the DTO into the domain model used by the service layer.
// TotalAmount is set to 0 — it will be confirmed by room-service after reservation.
func (r *CreateBookingRequest) ToModel() *model.Booking {
	return &model.Booking{
		User: model.User{
			UserID:         r.User.UserID,
			Name:           r.User.Name,
			Email:          r.User.Email,
			PhoneNumber:    r.User.PhoneNumber,
			NumberOfGuests: r.User.NumberOfGuests,
		},
		Room: model.Room{
			RoomID:     r.Room.RoomID,
			RoomNumber: r.Room.RoomNumber,
			Type:       r.Room.Type,
			Price:      r.Room.Price,
			Currency:   r.Room.Currency,
		},
		CheckInDate:  time.Time(r.CheckInDate),
		CheckOutDate: time.Time(r.CheckOutDate),
		TotalAmount:  0, //TODO: calculate
	}
}
