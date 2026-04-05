package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Booking collection in MongoDB
type Booking struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	Guest GuestSnapshot `bson:"guest" json:"guest"`
	Room  RoomSnapshot  `bson:"room" json:"room"`

	CheckInDate  time.Time `bson:"check_in_date" json:"check_in_date"`
	CheckOutDate time.Time `bson:"check_out_date" json:"check_out_date"`
	TotalAmount  float64   `bson:"total_amount" json:"total_amount"`
	Currency     string    `bson:"currency" json:"currency"`

	BookingStatus BookingStatus `bson:"status" json:"status"`                 // e.g., "confirmed", "cancelled"
	PaymentStatus PaymentStatus `bson:"payment_status" json:"payment_status"` // e.g., "paid", "pending"

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Snapshot data from User service
type GuestSnapshot struct {
	UserID string `bson:"user_id" json:"user_id"`
	Name   string `bson:"name" json:"name"`
	Email  string `bson:"email" json:"email"`
}

// Snapshot data from Room service
type RoomSnapshot struct {
	RoomID     string  `bson:"room_id" json:"room_id"`
	RoomNumber string  `bson:"room_number" json:"room_number"`
	Type       string  `bson:"type" json:"type"` //Standard, Deluxe, Suite
	Price      float64 `bson:"price" json:"price"`
}

// Booking status
type BookingStatus string

const (
	StatusReserved   BookingStatus = "reserved"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusCancelled  BookingStatus = "cancelled"
	StatusCheckedIn  BookingStatus = "checked_in"
	StatusCheckedOut BookingStatus = "checked_out"
	StatusNoShow     BookingStatus = "no_show"
)

func (s BookingStatus) IsValid() bool {
	switch s {
	case StatusReserved, StatusConfirmed, StatusCancelled, StatusCheckedIn, StatusCheckedOut, StatusNoShow:
		return true
	}
	return false
}

// Payment status
type PaymentStatus string

const (
	PaymentPending           PaymentStatus = "pending"
	PaymentCompleted         PaymentStatus = "completed"
	PaymentFailed            PaymentStatus = "failed"
	PaymentRefunded          PaymentStatus = "refunded"           // Full refund
	PaymentPartiallyRefunded PaymentStatus = "partially_refunded" // Partial refund. Eg: Guest cancels after check-in, so only refund for unused nights
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
