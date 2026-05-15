package kafka

import (
	"booking/booking-service/event"
	"booking/booking-service/model"
	"booking/booking-service/service"
	"context"
	"encoding/json"
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
		kgo.DisableAutoCommit(),
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
			log.Printf("Received message: %s\n", string(record.Value))

			if err := json.Unmarshal(record.Value, &reservationResult); err != nil {
				log.Printf("Failed to unmarshal message (skipping): %v\n", err)
				// Permanent failure — commit to avoid infinite redelivery (poison pill)
				// TODO: publish to dead-letter queue for investigation
				c.client.CommitRecords(ctx, record)
				return
			}

			data := reservationResult.Data
			log.Printf("Processing reservation result for booking %s (success=%v)\n", data.BookingID, data.Success)

			// Reservation failed — update status only
			if !data.Success {
				err := c.service.UpdateBookingStatus(ctx, data.BookingID, model.StatusReservationFailed)
				if err != nil {
					log.Printf("Failed to update booking status to RESERVATION_FAILED: %v\n", err)
					return // Do NOT commit — message will be redelivered
				}
				// DB write succeeded, commit offset
				if err := c.client.CommitRecords(ctx, record); err != nil {
					log.Printf("Failed to commit offset: %v\n", err)
				}
				return
			}

			// Reservation succeeded — update pricing and status to RESERVED
			var nightlyRates []model.NightlyRate
			for _, rate := range data.NightlyRates {
				nightlyRates = append(nightlyRates, model.NightlyRate{
					Date:     rate.Date,
					Price:    rate.Price,
					Currency: rate.Currency,
				})
			}

			err := c.service.ConfirmBookingReservation(ctx, data.BookingID, nightlyRates, data.TotalAmount, data.Room.Price, data.Room.Currency)
			if err != nil {
				log.Printf("Failed to confirm booking reservation: %v\n", err)
				return // Do NOT commit — message will be redelivered
			}
			// DB write succeeded, commit offset
			if err := c.client.CommitRecords(ctx, record); err != nil {
				log.Printf("Failed to commit offset: %v\n", err)
			}
		})
	}
}

// Close gracefully shuts down the consumer
func (c *RoomReservationConsumer) Close() {
	c.client.Close()
}
