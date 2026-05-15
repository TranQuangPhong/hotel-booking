package kafka //TODO: extract to common package for all services

var (
	BookingBrokerAddress = "localhost:9092"

	//Topics
	BookingCreatedTopic        = "booking_created_events"
	RoomReservationEventsTopic = "room_reservation_events"
	PaymentEventsTopic         = "payment_events"

	//Consumer group IDs
	//Prefix: "booking_svc" to avoid conflicts with other services in the same cluster
	RoomReservationEventsGroupID = "booking_svc_room_reservation_events"
	PaymentEventsGroupID         = "booking_svc_payment_events"
)
