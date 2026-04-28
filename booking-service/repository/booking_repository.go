package repository

import (
	"booking/booking-service/model"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(collection *mongo.Collection) *BookingRepository {
	return &BookingRepository{collection: collection}
}

var (
	ErrInvalidBookingID = errors.New("invalid booking ID format")
)

// Get booking by ID
func (r *BookingRepository) GetBookingByID(ctx context.Context, id string) (*model.Booking, error) {
	bookingID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidBookingID
	}
	var booking model.Booking
	err = r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: bookingID}}).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

// Get bookings by user ID
func (r *BookingRepository) GetBookingsByUserID(ctx context.Context, userID string) ([]*model.Booking, error) {
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "guest.user_id", Value: userID}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []*model.Booking
	if err = cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// Create new booking (reserve room)
func (r *BookingRepository) CreateBooking(ctx context.Context, booking *model.Booking) error {
	_, err := r.collection.InsertOne(ctx, booking)
	//TODO: to use outbox with CDC (next phase)
	return err
}
