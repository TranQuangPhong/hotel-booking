package dto

import "booking/room-service/model"

// CreateRoomRequest is the DTO for incoming create-room requests.
// It only exposes fields that the admin/operator is allowed to provide.
// Internal fields (ID, Status, CreatedAt, UpdatedAt) are set by the service/repository layer.
type CreateRoomRequest struct {
	RoomNumber string  `json:"room_number" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	BasePrice  float64 `json:"base_price" binding:"required,gt=0"`
	Currency   string  `json:"currency" binding:"required"`
}

// ToModel converts the DTO into the domain model used by the service layer.
// Status defaults to ACTIVE on creation.
func (r *CreateRoomRequest) ToModel() *model.Room {
	return &model.Room{
		RoomNumber: r.RoomNumber,
		Type:       model.RoomType(r.Type),
		Status:     model.StatusActive,
		BasePrice:  r.BasePrice,
		Currency:   r.Currency,
	}
}
