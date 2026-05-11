package kafka

import (
	"booking/booking-service/model"
	"booking/booking-service/service"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

type BookingConsumer struct {
	client  *kgo.Client
	service *service.BookingService
}

func NewBookingConsumer(brokers []string, groupID string, topics []string, s *service.BookingService) (*BookingConsumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
		// CRITICAL: Disable auto-commit. We only commit after MongoDB saves successfully.
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, err
	}
	return &BookingConsumer{client: cl, service: s}, nil
}

// Start kicks off the infinite polling loop.
// It runs in a goroutine until the context is cancelled.
func (c *BookingConsumer) Start(ctx context.Context) {
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
			var booking *model.Booking
			//Print record.value in json format for debugging
			log.Printf("Received message: %s\n", string(record.Value))

			// Unmarshal the JSON from Kong
			if err := json.Unmarshal(record.Value, &booking); err != nil {
				log.Printf("Failed to unmarshal message: %v\n", err)
				return // Skip this message and continue with the next one
			}
			// --- BUSINESS LOGIC HERE ---
			fmt.Println("Processed booking message:", booking)
			err := c.service.CreateBooking(ctx, booking)
			if err != nil {
				log.Printf("Failed to create booking: %v\n", err)
				return // If DB fails, we do NOT commit!
			}

			// 5. Commit the offset ONLY after successful processing
			err = c.client.CommitRecords(ctx, record)
			if err != nil {
				log.Printf("Failed to commit record: %v\n", err)
			} else {
				log.Printf("Successfully processed and committed message with offset %d\n", record.Offset)
			}
		})
	}
}

// Close gracefully shuts down the consumer
func (c *BookingConsumer) Close() {
	c.client.Close()
}
