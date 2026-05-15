package event //TODO: extract to common package for all services

type EventEnvelope[T any] struct {
	TraceID   string `json:"trace_id"`   // ID duy nhất để trace toàn bộ flow (correlation ID)
	EventType string `json:"event_type"` // Loại sự kiện (room_query, booking_request, payment_success, payment_failed...)
	Producer  string `json:"producer"`   // Tên service phát sinh event, có thể dùng để định tuyến hoặc phân tích log
	Timestamp string `json:"timestamp"`  // Thời điểm phát sinh event
	Data      T      `json:"data"`       // Nội dung chính của event, có thể là bất kỳ struct nào tùy theo eventType
}

// BookingCreatedMsg is the Kafka message published when a booking is created.
// The Room service consumes this to reserve the room.
// Uses string dates (ISO 8601) for cross-service portability.
type BookingCreatedMsg struct {
	BookingID    string `json:"booking_id"`
	User         User   `json:"user"`
	RoomID       string `json:"room_id"`
	CheckInDate  string `json:"check_in_date"`  // Date only: "2006-01-02"
	CheckOutDate string `json:"check_out_date"` // Date only: "2006-01-02"
	CreatedAt    string `json:"created_at"`     // RFC 3339
}

type User struct {
	UserID         string `json:"user_id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	PhoneNumber    string `json:"phone_number"`
	NumberOfGuests int    `json:"number_of_guests"`
}

type ReservationResultMsg struct {
	BookingID    string        `json:"booking_id"`
	Success      bool          `json:"success"`
	ErrorCode    string        `json:"error_code,omitempty"` // Error code if success=false
	Reason       string        `json:"reason,omitempty"`     // Failure reason if success=false
	Room         Room          `json:"room"`
	User         User          `json:"user"`
	CheckInDate  string        `json:"check_in_date"`  // Date only: "2006-01-02"
	CheckOutDate string        `json:"check_out_date"` // Date only: "2006-01-02"
	NightlyRates []NightlyRate `json:"nightly_rates"`  // Per-night price breakdown from inventory slots
	TotalAmount  float64       `json:"total_amount"`   // Sum of nightly rates, calculated by room-service
	CreatedAt    string        `json:"created_at"`     // RFC 3339
}

// NightlyRate represents the confirmed price for a single night from inventory.
type NightlyRate struct {
	Date     string  `json:"date"`     // Date only: "2006-01-02"
	Price    float64 `json:"price"`    // Actual slot price for this night
	Currency string  `json:"currency"` // Currency for this night's price
}

type Room struct {
	RoomID     string  `json:"room_id"`
	RoomNumber string  `json:"room_number"`
	Type       string  `json:"type"` //Standard, Deluxe, Suite
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
}
