package repository

import (
	"booking/room-service/model"
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type InventoryRepository struct {
	collection *mongo.Collection
}

func NewInventoryRepository(c *mongo.Collection) *InventoryRepository {
	return &InventoryRepository{collection: c}
}

var (
	ErrInventoryNotFound = errors.New("inventory document not found")
)

func (r *InventoryRepository) CreateInventory(ctx context.Context, inventoryBucket []model.Inventory) error {
	_, err := r.collection.InsertMany(ctx, inventoryBucket)
	return err
}

// ReserveSlots sets status=RESERVED and booking_id for the given days within a single month's inventory document.
// For cross-month reservations, the service layer must call this once per month.
func (r *InventoryRepository) ReserveSlots(ctx context.Context, roomID string, month string, days []int, bookingID string) error {
	filter := bson.D{
		{Key: "room_id", Value: roomID},
		{Key: "month", Value: month},
	}

	setFields := bson.D{}
	for _, day := range days {
		prefix := fmt.Sprintf("slots.%d", day)
		setFields = append(setFields,
			bson.E{Key: prefix + ".status", Value: model.StatusReserved},
			bson.E{Key: prefix + ".booking_id", Value: bookingID},
		)
	}

	update := bson.D{{Key: "$set", Value: setFields}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrInventoryNotFound
	}
	return nil
}

// GetSlotsByDays retrieves specific slots from a single month's inventory document.
// Returns a map of day -> Slot for the requested days.
// For cross-month lookups, the service layer must call this once per month.
func (r *InventoryRepository) GetSlotsByDays(ctx context.Context, roomID string, month string, days []int) (map[int]model.Slot, error) {
	filter := bson.D{
		{Key: "room_id", Value: roomID},
		{Key: "month", Value: month},
	}

	var inventory model.Inventory
	err := r.collection.FindOne(ctx, filter).Decode(&inventory)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrInventoryNotFound
		}
		return nil, err
	}

	result := make(map[int]model.Slot, len(days))
	for _, day := range days {
		key := fmt.Sprintf("%d", day)
		if slot, ok := inventory.Slots[key]; ok {
			result[day] = slot
		}
	}

	return result, nil
}
