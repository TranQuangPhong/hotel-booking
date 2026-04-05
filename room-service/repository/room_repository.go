package repository

import (
	"booking/room-service/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RoomRepository struct {
	collection *mongo.Collection
}

func NewRoomRepository(collection *mongo.Collection) *RoomRepository {
	return &RoomRepository{collection: collection}
}

// Get all rooms
func (r *RoomRepository) GetAllRooms(ctx context.Context) ([]*model.Room, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []*model.Room
	if err = cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}

	return rooms, nil
}

// Get room by ID
func (r *RoomRepository) GetRoomByID(ctx context.Context, id string) (*model.Room, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var room model.Room
	err = r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: objID}}).Decode(&room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

// Insert document
func (r *RoomRepository) CreateRoom(ctx context.Context, room *model.Room) error {
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	room.Status = "available"

	_, err := r.collection.InsertOne(ctx, room)
	return err
}

// Update room status
func (r *RoomRepository) UpdateRoomStatus(ctx context.Context, id string, status model.RoomStatus) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "_id", Value: objID}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: status},
			{Key: "updated_at", Value: time.Now()},
		}},
	}
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err

}
