package service

import (
	"booking/booking-service/model"
	"booking/booking-service/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

type BookingRepository interface {
	GetBookingByID(ctx context.Context, id string) (*model.Booking, error)
	GetBookingsByUserID(ctx context.Context, userID string) ([]*model.Booking, error)
	CreateBooking(ctx context.Context, booking *model.Booking) (string, error)
	UpdateBookingStatus(ctx context.Context, bookingID string, status model.BookingStatus) error
	UpdateBookingPricing(ctx context.Context, bookingID string, nightlyRates []model.NightlyRate, totalAmount float64, roomPrice float64, roomCurrency string) error
}

type BookingService struct {
	bookingRepository BookingRepository
}

func NewBookingService(repo BookingRepository) *BookingService {
	return &BookingService{bookingRepository: repo}
}

var (
	ErrBookingStatusInvalid = errors.New("booking status invalid")
)

func (s *BookingService) GetBookingByID(ctx context.Context, id string) (*model.Booking, error) {
	booking, err := s.bookingRepository.GetBookingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidBookingID) {
			return nil, nil //Hide internal error details, just return nil to indicate not found
		}
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

// Create temp booking immediately
func (s *BookingService) CreateBooking(ctx context.Context, booking *model.Booking) (string, error) {

	booking.CreatedAt = time.Now().UTC()
	booking.UpdatedAt = time.Now().UTC()
	booking.Status = model.StatusPending
	booking.PaymentStatus = model.PaymentPending

	bookingID, error := s.bookingRepository.CreateBooking(ctx, booking)
	if error != nil {
		return "", error
	}
	return bookingID, nil
}

func (s *BookingService) UpdateBookingStatus(ctx context.Context, bookingID string, status model.BookingStatus) error {
	if status.IsValid() == false {
		return ErrBookingStatusInvalid
	}
	return s.bookingRepository.UpdateBookingStatus(ctx, bookingID, status)
}

// ConfirmBookingReservation updates the booking with confirmed pricing from room-service and sets status to RESERVED.
func (s *BookingService) ConfirmBookingReservation(ctx context.Context, bookingID string, nightlyRates []model.NightlyRate, totalAmount float64, roomPrice float64, roomCurrency string) error {
	// Update pricing and room price/currency
	err := s.bookingRepository.UpdateBookingPricing(ctx, bookingID, nightlyRates, totalAmount, roomPrice, roomCurrency)
	if err != nil {
		return fmt.Errorf("failed to update booking pricing: %w", err)
	}
	// Update status to RESERVED
	err = s.bookingRepository.UpdateBookingStatus(ctx, bookingID, model.StatusReserved)
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}
	return nil
}
