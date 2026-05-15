package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Bookings collection in MongoDB
type Booking struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	User User `bson:"user" json:"user"`
	Room Room `bson:"room" json:"room"`

	CheckInDate  time.Time `bson:"check_in_date" json:"check_in_date"`
	CheckOutDate time.Time `bson:"check_out_date" json:"check_out_date"`
	TotalAmount  float64   `bson:"total_amount" json:"total_amount"` // Confirmed by room-service, 0 until reservation confirmed

	NightlyRates []NightlyRate `bson:"nightly_rates,omitempty" json:"nightly_rates,omitempty"` // Per-night price breakdown from room-service

	Status        BookingStatus `bson:"status" json:"status"`                 // e.g., "booked", "cancelled"
	PaymentStatus PaymentStatus `bson:"payment_status" json:"payment_status"` // e.g., "paid", "pending"

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Snapshot data from User service
type User struct {
	UserID         string `bson:"user_id" json:"user_id"`
	Name           string `bson:"name" json:"name"`
	Email          string `bson:"email" json:"email"`
	PhoneNumber    string `bson:"phone_number" json:"phone_number"`
	NumberOfGuests int    `bson:"number_of_guests" json:"number_of_guests"`
}

// Snapshot data from Room service
type Room struct {
	RoomID     string  `bson:"room_id" json:"room_id"`
	RoomNumber string  `bson:"room_number" json:"room_number"`
	Type       string  `bson:"type" json:"type"` //Standard, Deluxe, Suite
	Price      float64 `bson:"price" json:"price"`
	Currency   string  `bson:"currency" json:"currency"`
}

// NightlyRate represents the confirmed price for a single night, provided by room-service.
type NightlyRate struct {
	Date     string  `bson:"date" json:"date"`         // Date only: "2006-01-02"
	Price    float64 `bson:"price" json:"price"`       // Actual slot price for this night
	Currency string  `bson:"currency" json:"currency"` // Currency for this night's price
}

// Booking status
type BookingStatus string

const (
	StatusPending           BookingStatus = "PENDING"
	StatusReserved          BookingStatus = "RESERVED"
	StatusReservationFailed BookingStatus = "RESERVATION_FAILED"
	StatusBooked            BookingStatus = "BOOKED"
	StatusCancelled         BookingStatus = "CANCELLED"
	StatusCheckedIn         BookingStatus = "CHECKED_IN"
	StatusCheckedOut        BookingStatus = "CHECKED_OUT"
	StatusNoShow            BookingStatus = "NO_SHOW"
)

func (s BookingStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusReserved, StatusReservationFailed, StatusBooked, StatusCancelled, StatusCheckedIn, StatusCheckedOut, StatusNoShow:
		return true
	}
	return false
}

// Payment status
type PaymentStatus string

const (
	PaymentPending           PaymentStatus = "PENDING"
	PaymentCompleted         PaymentStatus = "COMPLETED"
	PaymentFailed            PaymentStatus = "FAILED"
	PaymentRefunded          PaymentStatus = "REFUNDED"           // Full refund
	PaymentPartiallyRefunded PaymentStatus = "PARTIALLY_REFUNDED" // Partial refund. Eg: user cancels after check-in, so only refund for unused nights
)

func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentPending, PaymentCompleted, PaymentFailed, PaymentRefunded, PaymentPartiallyRefunded:
		return true
	}
	return false
}

func (s PaymentStatus) IsTerminal() bool {
	return s == PaymentRefunded || s == PaymentFailed
}
