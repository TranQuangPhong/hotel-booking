package repository

import (
	"booking/booking-service/model"
	"context"
	"errors"
	"time"

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
	ErrBookingNotFound  = errors.New("booking not found")
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
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "user.user_id", Value: userID}})
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
func (r *BookingRepository) CreateBooking(ctx context.Context, booking *model.Booking) (string, error) {
	result, err := r.collection.InsertOne(ctx, booking)
	if err != nil {
		return "", err
	}

	oid, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return "", errors.New("failed to extract inserted ID")
	}

	return oid.Hex(), nil
}

func (r *BookingRepository) UpdateBookingStatus(ctx context.Context, bookingID string, status model.BookingStatus) error {
	id, err := bson.ObjectIDFromHex(bookingID)
	if err != nil {
		return ErrInvalidBookingID
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: status},
			{Key: "updated_at", Value: time.Now().UTC()},
		}},
	}
	result, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBookingNotFound
	}
	return nil
}

// UpdateBookingPricing sets the confirmed nightly rates, total amount, and room price/currency from room-service.
func (r *BookingRepository) UpdateBookingPricing(ctx context.Context, bookingID string, nightlyRates []model.NightlyRate, totalAmount float64, roomPrice float64, roomCurrency string) error {
	id, err := bson.ObjectIDFromHex(bookingID)
	if err != nil {
		return ErrInvalidBookingID
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "nightly_rates", Value: nightlyRates},
			{Key: "total_amount", Value: totalAmount},
			{Key: "room.price", Value: roomPrice},
			{Key: "room.currency", Value: roomCurrency},
			{Key: "updated_at", Value: time.Now().UTC()},
		}},
	}
	result, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBookingNotFound
	}
	return nil
}
