package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Room collection in MongoDB
type Room struct {
	ID         bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	RoomNumber string           `bson:"room_number" json:"room_number"`
	Type       RoomType         `bson:"type" json:"type"`             //Standard, Deluxe, Suite
	Status     RoomMasterStatus `bson:"status" json:"status"`         //Master status
	BasePrice  float64          `bson:"base_price" json:"base_price"` //For initialization only. Inventory price overrides this.
	Currency   string           `bson:"currency" json:"currency"`     //USD, EUR, etc
	CreatedAt  time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time        `bson:"updated_at" json:"updated_at"`
}

type RoomMasterStatus string
type RoomType string

const (
	StatusActive   RoomMasterStatus = "ACTIVE"   // Room is open for business
	StatusInactive RoomMasterStatus = "INACTIVE" // Out of order indefinitely
	StatusArchived RoomMasterStatus = "ARCHIVED" // Physically removed/deleted
)

const (
	TypeStandard RoomType = "STANDARD"
	TypeDeluxe   RoomType = "DELUXE"
	TypeSuite    RoomType = "SUITE"
)

func (t RoomType) IsValid() bool {
	switch t {
	case TypeStandard, TypeDeluxe, TypeSuite:
		return true
	}
	return false
}

func (s RoomMasterStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusArchived:
		return true
	}
	return false
}
