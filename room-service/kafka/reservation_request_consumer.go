package kafka

import (
	"booking/room-service/event"
	"booking/room-service/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// TODO: Ordering - idempotency - exactly-once semantics
type ReservationConsumer struct {
	client   *kgo.Client
	service  *service.RoomService
	producer *ReservationProducer
}

func NewReservationConsumer(brokers []string, groupID string, topics []string, s *service.RoomService, p *ReservationProducer) (*ReservationConsumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
	)
	if err != nil {
		return nil, err
	}

	return &ReservationConsumer{client: cl, service: s, producer: p}, nil
}

func (c *ReservationConsumer) Start(ctx context.Context) {
	for {
		// 1. Fetch a batch of messages (blocks until messages are available or ctx is cancelled)
		fetches := c.client.PollFetches(ctx)
		// 2. Check for shutdown signals
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return
		}
		// 3. Handle consumer errors (like losing connection to the broker)
		fetches.EachError(func(topic string, partition int32, err error) {
			log.Printf("Fetch error on topic %s, partition %d: %v\n", topic, partition, err)
		})

		// 4. Process the actual messages
		var bookingCreatedMsg *event.EventEnvelope[event.BookingCreatedMsg]
		fetches.EachRecord(func(record *kgo.Record) {
			//Print record.value in json format for debugging
			log.Printf("Received message: %s\n", string(record.Value))
			//Unmarshal JSON msg
			if err := json.Unmarshal(record.Value, &bookingCreatedMsg); err != nil {
				log.Printf("Failed to unmarshal message: %v\n", err)
				return // Skip this message and continue with the next one
			}

			// --- BUSINESS LOGIC HERE ---
			fmt.Println("Processed bookingCreatedMsg message:", bookingCreatedMsg)
			//TODO
			//Check room status
			//Reserve room if available

			//Publish reservation result
			msg := event.EventEnvelope[event.ReservationResultMsg]{
				TraceID:   "0", //TODO: generate real trace ID for distributed tracing
				EventType: "reservation_result",
				Producer:  "room-service",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Data: event.ReservationResultMsg{
					BookingID: bookingCreatedMsg.Data.BookingID,
					Success:   true, //TODO: set real reservation result
					// ErrorCode:        "ROOM_UNAVAILABLE", //TODO: set real error code if reservation failed
					// Reason:           "Room is already booked for the selected dates", //TODO: set real reason if reservation failed
					Room: event.Room{
						RoomID:     bookingCreatedMsg.Data.RoomID,
						RoomNumber: "101",
						Type:       "Standard",
						Price:      100.0,
						Currency:   "USD",
					},
					User: event.User{
						UserID:         bookingCreatedMsg.Data.User.UserID,
						Name:           bookingCreatedMsg.Data.User.Name,
						Email:          bookingCreatedMsg.Data.User.Email,
						PhoneNumber:    bookingCreatedMsg.Data.User.PhoneNumber,
						NumberOfGuests: bookingCreatedMsg.Data.User.NumberOfGuests,
					},
					CheckInDate:  bookingCreatedMsg.Data.CheckInDate,
					CheckOutDate: bookingCreatedMsg.Data.CheckOutDate,
					TotalAmount:  100.0, //TODO: calculate real total amount
					Currency:     "USD",
					CreatedAt:    time.Now().UTC().Format(time.RFC3339), //TODO: set real created at time
				},
			}
			producerCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if err := c.producer.PublishReservationResult(producerCtx, msg); err != nil {
				log.Printf("Failed to publish reservation result: %v\n", err)
				return
			}
		})
	}
}

// Close gracefully shuts down the consumer
func (c *ReservationConsumer) Close() {
	c.client.Close()
}
