package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Inventory struct {
	ID        bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	RoomID    string          `bson:"room_id" json:"room_id"` // Matches Room.ID.Hex()
	Month     string          `bson:"month" json:"month"`     // Format: YYYY-MM
	Slots     map[string]Slot `bson:"slots" json:"slots"`     // Key: day of month (1-31), Value: Slot struct
	CreatedAt time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time       `bson:"updated_at" json:"updated_at"`
}

type Slot struct {
	Status    RoomStatus `bson:"status" json:"status"`
	Price     float64    `bson:"price" json:"price"`       // Daily override price
	Currency  string     `bson:"currency" json:"currency"` // Currency for this slot's price
	BookingID string     `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
}

type RoomStatus string

const (
	StatusAvailable RoomStatus = "AVAILABLE"
	// PENDING status for rooms that are in the process of being booked but not yet confirmed..
	// This can help prevent race conditions where multiple users try to book the same room at the same time.
	StatusReserved    RoomStatus = "RESERVED"
	StatusBooked      RoomStatus = "BOOKED"
	StatusMaintenance RoomStatus = "MAINTENANCE"
)

func (s RoomStatus) IsValid() bool {
	switch s {
	case StatusAvailable, StatusBooked, StatusMaintenance, StatusReserved:
		return true
	}
	return false
}
