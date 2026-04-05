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
)

func (s RoomStatus) IsValid() bool {
	switch s {
	case StatusAvailable, StatusBooked, StatusMaintenance:
		return true
	}
	return false
}

// Kafka event payload
type KafkaEvent struct {
	TraceID   string       `json:"traceId"`
	EventType string       `json:"eventType"`
	Timestamp time.Time    `json:"timestamp"`
	User      UserBlock    `json:"userBlock"`
	Payload   PayloadBlock `json:"payloadBlock"`
	Saga      SagaBlock    `json:"sagaBlock"`
}

type SagaBlock struct {
	Step         string `json:"step"`
	Status       string `json:"status"`       //pending, completed, failed
	Compensation string `json:"compensation"` //compensation action if failed
}

type UserBlock struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
	Email  string   `json:"email"`
}

type PayloadBlock struct {
	EventID  string  `json:"eventId"`
	RoomID   string  `json:"roomId"`
	CheckIn  string  `json:"checkIn"`
	CheckOut string  `json:"checkOut"`
	Guests   int     `json:"guests"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}
