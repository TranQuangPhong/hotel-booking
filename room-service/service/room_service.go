package service

import (
	"booking/room-service/model"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RoomRepository interface {
	GetAllRooms(ctx context.Context) ([]*model.Room, error)
	GetRoomByID(ctx context.Context, id string) (*model.Room, error)
	CreateRoom(ctx context.Context, room *model.Room) (string, error)
	UpdateRoomStatus(ctx context.Context, id string, status model.RoomMasterStatus) error
}

type InventoryRepository interface {
	CreateInventory(ctx context.Context, inventoryBucket []model.Inventory) error
	ReserveSlots(ctx context.Context, roomID string, month string, days []int, bookingID string) error
	GetSlotsByDays(ctx context.Context, roomID string, month string, days []int) (map[int]model.Slot, error)
}

type RoomService struct {
	client              *mongo.Client
	roomRepository      RoomRepository
	inventoryRepository InventoryRepository
}

func NewRoomService(cl *mongo.Client, roomRepo RoomRepository, inventoryRepo InventoryRepository) *RoomService {
	return &RoomService{client: cl, roomRepository: roomRepo, inventoryRepository: inventoryRepo}
}

var (
	ErrRoomNotFound          = errors.New("room not found")
	ErrInvalidRoomStatus     = errors.New("provided room status is invalid")
	ErrMaintenanceNotAllowed = errors.New("cannot move to maintenance while room is booked") //TODO: use when inventory status check is implemented
)

// Get all rooms
func (s *RoomService) GetAllRooms(ctx context.Context) ([]*model.Room, error) {
	rooms, err := s.roomRepository.GetAllRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rooms: %w", err)
	}
	if rooms == nil {
		return nil, ErrRoomNotFound
	}

	return rooms, nil
}

// Get room by ID
func (s *RoomService) GetRoomByID(ctx context.Context, id string) (*model.Room, error) {
	room, err := s.roomRepository.GetRoomByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch room: %w", err)
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// Transaction: Create new room with inventory for next 12 months
func (s *RoomService) CreateRoomWithInventory(ctx context.Context, room *model.Room) error {
	//Start session
	session, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)
	//Start transaction
	_, err = session.WithTransaction(ctx, func(sessionCtx context.Context) (any, error) {
		//Create room
		roomID, err := s.roomRepository.CreateRoom(sessionCtx, room)
		if err != nil {
			return nil, fmt.Errorf("failed to create room: %w", err)
		}

		//Create inventory for next 12 months
		now := time.Now().UTC()
		inventoryBucket := make([]model.Inventory, 0, 12)

		for i := 1; i <= 12; i++ {
			// Calculate the target month
			targetMonth := now.AddDate(0, i, 0)
			year, month, _ := targetMonth.Date()

			// Get actual number of days in this month
			daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

			// Build slots map with dynamic days
			slots := make(map[string]model.Slot, daysInMonth)
			for day := 1; day <= daysInMonth; day++ {
				key := fmt.Sprintf("%d", day)
				slots[key] = model.Slot{
					Status:   model.StatusAvailable,
					Price:    room.BasePrice,
					Currency: room.Currency,
				}
			}

			inventory := model.Inventory{
				RoomID:    roomID,
				Month:     targetMonth.Format("2006-01"),
				Slots:     slots,
				CreatedAt: now,
				UpdatedAt: now,
			}
			inventoryBucket = append(inventoryBucket, inventory)
		}

		err = s.inventoryRepository.CreateInventory(sessionCtx, inventoryBucket)
		if err != nil {
			return nil, fmt.Errorf("failed to create inventory: %w", err)
		}

		return nil, nil
	})

	return err
}

// Update room master status
func (s *RoomService) UpdateRoomStatus(ctx context.Context, id string, status model.RoomMasterStatus) error {
	if !status.IsValid() {
		return ErrInvalidRoomStatus
	}

	currentRoom, err := s.roomRepository.GetRoomByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch room: %w", err)
	}
	if currentRoom == nil {
		return ErrRoomNotFound
	}

	return s.roomRepository.UpdateRoomStatus(ctx, id, status)
}

// ReservationResult holds the outcome of a room reservation attempt.
type ReservationResult struct {
	Success      bool
	ErrorCode    string
	Reason       string
	NightlyRates []NightlyRate
	TotalAmount  float64
}

// NightlyRate represents the confirmed price for a single night.
type NightlyRate struct {
	Date     string
	Price    float64
	Currency string
}

// ReserveRoom checks availability, calculates pricing from inventory slots, and reserves the room.
// Handles cross-month date ranges by splitting into per-month operations.
// Returns a ReservationResult with nightly breakdown and total, or failure details.
func (s *RoomService) ReserveRoom(ctx context.Context, roomID string, checkInDate string, checkOutDate string, bookingID string) (*ReservationResult, error) {
	// Parse dates
	checkIn, err := time.Parse(time.DateOnly, checkInDate)
	if err != nil {
		return nil, fmt.Errorf("invalid check_in_date format: %w", err)
	}
	checkOut, err := time.Parse(time.DateOnly, checkOutDate)
	if err != nil {
		return nil, fmt.Errorf("invalid check_out_date format: %w", err)
	}

	// Split date range into per-month day groups
	monthDays := splitDateRange(checkIn, checkOut)

	// Step 1: Check availability
	// TODO: Check room availability on Redis before reserving slots

	// Step 2: Reserve all slots and collect prices
	var nightlyRates []NightlyRate
	var totalAmount float64

	for month, days := range monthDays {
		// Reserve slots
		err := s.inventoryRepository.ReserveSlots(ctx, roomID, month, days, bookingID)
		if err != nil {
			// TODO: compensate previously reserved months on failure
			return nil, fmt.Errorf("failed to reserve slots for month %s: %w", month, err)
		}

		// Collect prices from reserved slots
		slots, err := s.inventoryRepository.GetSlotsByDays(ctx, roomID, month, days)
		if err != nil {
			return nil, fmt.Errorf("failed to get slot prices for month %s: %w", month, err)
		}

		for _, day := range days {
			slot, exists := slots[day]
			if !exists {
				continue
			}
			year := month[:4]
			monthNum := month[5:]
			dateStr := fmt.Sprintf("%s-%s-%02d", year, monthNum, day)
			nightlyRates = append(nightlyRates, NightlyRate{
				Date:     dateStr,
				Price:    slot.Price,
				Currency: slot.Currency,
			})
			totalAmount += slot.Price
		}
	}

	return &ReservationResult{
		Success:      true,
		NightlyRates: nightlyRates,
		TotalAmount:  totalAmount,
	}, nil
}

// splitDateRange groups dates from checkIn (inclusive) to checkOut (exclusive) by month.
// Returns a map of "YYYY-MM" -> []int (day numbers).
func splitDateRange(checkIn, checkOut time.Time) map[string][]int {
	result := make(map[string][]int)

	current := checkIn
	for current.Before(checkOut) {
		month := current.Format("2006-01")
		day := current.Day()
		result[month] = append(result[month], day)
		current = current.AddDate(0, 0, 1)
	}

	return result
}
