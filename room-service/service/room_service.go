package service

import (
	"booking/room-service/model"
	"context"
	"errors"
	"fmt"
)

type RoomRepository interface {
	GetAllRooms(ctx context.Context) ([]*model.Room, error)
	GetRoomByID(ctx context.Context, id string) (*model.Room, error)
	CreateRoom(ctx context.Context, room *model.Room) error
	UpdateRoomStatus(ctx context.Context, id string, status model.RoomStatus) error
}

type RoomService struct {
	repository RoomRepository
}

func NewRoomService(repo RoomRepository) *RoomService {
	return &RoomService{repository: repo}
}

var (
	ErrRoomNotFound          = errors.New("room not found")
	ErrInvalidRoomStatus     = errors.New("provided room status is invalid")
	ErrMaintenanceNotAllowed = errors.New("cannot move to maintenance while room is booked")
)

// Get all rooms
func (s *RoomService) GetAllRooms(ctx context.Context) ([]*model.Room, error) {
	rooms, err := s.repository.GetAllRooms(ctx)
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
	room, err := s.repository.GetRoomByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch room: %w", err)
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// Create new room
func (s *RoomService) CreateRoom(ctx context.Context, room *model.Room) error {
	return s.repository.CreateRoom(ctx, room)
}

// Update room status
func (s *RoomService) UpdateRoomStatus(ctx context.Context, id string, status model.RoomStatus) error {
	if !status.IsValid() {
		return ErrInvalidRoomStatus
	}

	currentRoom, err := s.repository.GetRoomByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch room: %w", err)
	}
	if currentRoom == nil {
		return ErrRoomNotFound
	}

	if status == model.StatusMaintenance && currentRoom.Status == model.StatusBooked {
		return ErrMaintenanceNotAllowed
	}

	return s.repository.UpdateRoomStatus(ctx, id, status)
}
