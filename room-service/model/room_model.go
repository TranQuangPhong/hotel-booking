package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Room collection in MongoDB
type Room struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomNumber string        `bson:"room_number" json:"room_number"`
	Type       string        `bson:"type" json:"type"` //Standard, Deluxe, Suite
	Price      float64       `bson:"price" json:"price"`
	Currency   string        `bson:"currency" json:"currency"` //USD, EUR, etc.
	Status     RoomStatus    `bson:"status" json:"status"`     //available, booked, maintenance
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time     `bson:"updated_at" json:"updated_at"`
}

type RoomStatus string

const (
	StatusAvailable   RoomStatus = "available"
	StatusBooked      RoomStatus = "booked"
	StatusMaintenance RoomStatus = "maintenance"
	// PENDING status for rooms that are in the process of being booked but not yet confirmed..
	// This can help prevent race conditions where multiple users try to book the same room at the same time.
	StatusPendingReservation RoomStatus = "pending_reservation" //TODO: consider remove
)

func (s RoomStatus) IsValid() bool {
	switch s {
	case StatusAvailable, StatusBooked, StatusMaintenance, StatusPendingReservation:
		return true
	}
	return false
}
