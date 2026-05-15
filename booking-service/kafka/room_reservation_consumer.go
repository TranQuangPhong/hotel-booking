package kafka

import (
	"booking/booking-service/event"
	"booking/booking-service/model"
	"booking/booking-service/service"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

// TODO: Ordering - idempotency - exactly-once semantics
type RoomReservationConsumer struct {
	client  *kgo.Client
	service *service.BookingService
}

func NewRoomReservationConsumer(brokers []string, groupID string, topics []string, s *service.BookingService) (*RoomReservationConsumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
	)
	if err != nil {
		return nil, err
	}
	return &RoomReservationConsumer{client: cl, service: s}, nil
}

// Start kicks off the infinite polling loop.
// It runs in a goroutine until the context is cancelled.
func (c *RoomReservationConsumer) Start(ctx context.Context) {
	for {
		// 1. Fetch a batch of messages (blocks until messages are available or ctx is cancelled)
		fetches := c.client.PollFetches(ctx)
		// 2. Check for shutdown signals
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return // Exit the loop if the client is closed or context is cancelled
		}
		// 3. Handle consumer errors (like losing connection to the broker)
		fetches.EachError(func(topic string, partition int32, err error) {
			log.Printf("Fetch error on topic %s partition %d: %v\n", topic, partition, err)
		})

		// 4. Process the actual messages
		fetches.EachRecord(func(record *kgo.Record) {
			var reservationResult *event.EventEnvelope[event.ReservationResultMsg]
			//Print record.value in json format for debugging
			log.Printf("Received message: %s\n", string(record.Value))

			// Unmarshal JSON msg
			if err := json.Unmarshal(record.Value, &reservationResult); err != nil {
				log.Printf("Failed to unmarshal message: %v\n", err)
				return // Skip this message and continue with the next one
			}
			// --- BUSINESS LOGIC HERE ---
			// Update booking status based on reservation result
			fmt.Println("Processed reservation result message:", reservationResult.Data)
			var status model.BookingStatus
			if reservationResult.Data.Success == false {
				status = model.StatusReservationFailed
			} else {
				status = model.StatusReserved
			}
			err := c.service.UpdateBookingStatus(ctx, reservationResult.Data.BookingID, status)
			if err != nil {
				log.Printf("Failed to update booking: %v\n", err)
			}
		})
	}
}

// Close gracefully shuts down the consumer
func (c *RoomReservationConsumer) Close() {
	c.client.Close()
}
