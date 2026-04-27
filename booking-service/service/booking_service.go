package service

import (
	"booking/booking-service/model"
	"context"
	"fmt"
	"time"
)

type BookingRepository interface {
	GetBookingByID(ctx context.Context, id string) (*model.Booking, error)
	GetBookingsByUserID(ctx context.Context, userID string) ([]*model.Booking, error)
	CreateBooking(ctx context.Context, booking *model.Booking) error
}

type BookingService struct {
	bookingRepository BookingRepository
}

func NewBookingService(repo BookingRepository) *BookingService {
	return &BookingService{bookingRepository: repo}
}

func (s *BookingService) GetBookingByID(ctx context.Context, id string) (*model.Booking, error) {
	booking, err := s.bookingRepository.GetBookingByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking by ID: %w", err)
	}
	return booking, nil
}

func (s *BookingService) GetBookingsByUserID(ctx context.Context, userID string) ([]*model.Booking, error) {
	bookings, err := s.bookingRepository.GetBookingsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings by user ID: %w", err)
	}
	return bookings, nil
}

func (s *BookingService) CreateBooking(ctx context.Context, booking *model.Booking) error {

	//TODO: check room availability before creating booking

	booking.CreatedAt = time.Now()
	booking.UpdatedAt = time.Now()
	booking.BookingStatus = model.StatusReserved

	error := s.bookingRepository.CreateBooking(ctx, booking)
	if error != nil {
		return error
	}
	//TODO: Send msg to Kafka
	return nil
}
