package kafka

import (
	"booking/room-service/event"
	"booking/room-service/model"
	"booking/room-service/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// TODO: Ordering - idempotency - exactly-once semantics
type ReservationRequestConsumer struct {
	client   *kgo.Client
	service  *service.RoomService
	producer *ReservationProducer
}

func NewReservationConsumer(brokers []string, groupID string, topics []string, s *service.RoomService, p *ReservationProducer) (*ReservationRequestConsumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, err
	}

	return &ReservationRequestConsumer{client: cl, service: s, producer: p}, nil
}

func (c *ReservationRequestConsumer) Start(ctx context.Context) {
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
		fetches.EachRecord(func(record *kgo.Record) {
			var bookingCreatedMsg *event.EventEnvelope[event.BookingCreatedMsg]
			log.Printf("Received message: %s\n", string(record.Value))

			if err := json.Unmarshal(record.Value, &bookingCreatedMsg); err != nil {
				log.Printf("Failed to unmarshal message (skipping): %v\n", err)
				// Permanent failure — commit to avoid infinite redelivery (poison pill)
				// TODO: publish to dead-letter queue for investigation
				c.client.CommitRecords(ctx, record)
				return
			}

			data := bookingCreatedMsg.Data
			log.Printf("Processing booking reservation: %s\n", data.BookingID)

			// Reserve room and calculate pricing from inventory
			result, err := c.service.ReserveRoom(ctx, data.RoomID, data.CheckInDate, data.CheckOutDate, data.BookingID)
			if err != nil {
				log.Printf("Failed to reserve room: %v\n", err)
				// Publish failure so booking-service can update status
				c.publishFailure(ctx, bookingCreatedMsg, "INTERNAL_ERROR", fmt.Sprintf("failed to reserve room: %v", err))
				return // Do NOT commit — message will be redelivered
			}

			// Fetch room details for the response
			room, err := c.service.GetRoomByID(ctx, data.RoomID)
			if err != nil {
				log.Printf("Failed to fetch room details: %v\n", err)
				// Publish failure so booking-service can update status
				c.publishFailure(ctx, bookingCreatedMsg, "INTERNAL_ERROR", fmt.Sprintf("failed to fetch room details: %v", err))
				return // Do NOT commit — message will be redelivered
			}

			// Build nightly rates for the event message
			var eventRates []event.NightlyRate
			for _, rate := range result.NightlyRates {
				eventRates = append(eventRates, event.NightlyRate{
					Date:     rate.Date,
					Price:    rate.Price,
					Currency: rate.Currency,
				})
			}

			// Publish success result
			c.publishSuccess(ctx, bookingCreatedMsg, result, room, eventRates)
			log.Printf("Reservation result published for booking %s (success=%v)\n", data.BookingID, result.Success)

			// DB write succeeded + result published, commit offset
			if err := c.client.CommitRecords(ctx, record); err != nil {
				log.Printf("Failed to commit offset: %v\n", err)
			}
		})
	}
}

// Close gracefully shuts down the consumer
func (c *ReservationRequestConsumer) Close() {
	c.client.Close()
}

// publishSuccess sends a success ReservationResultMsg with pricing and room details.
func (c *ReservationRequestConsumer) publishSuccess(ctx context.Context, msg *event.EventEnvelope[event.BookingCreatedMsg], result *service.ReservationResult, room *model.Room, eventRates []event.NightlyRate) {
	data := msg.Data
	successMsg := event.EventEnvelope[event.ReservationResultMsg]{
		TraceID:   msg.TraceID,
		EventType: "reservation_result",
		Producer:  "room-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: event.ReservationResultMsg{
			BookingID: data.BookingID,
			Success:   true,
			Room: event.Room{
				RoomID:     room.ID.Hex(),
				RoomNumber: room.RoomNumber,
				Type:       string(room.Type),
				Price:      room.BasePrice,
				Currency:   room.Currency,
			},
			User: event.User{
				UserID:         data.User.UserID,
				Name:           data.User.Name,
				Email:          data.User.Email,
				PhoneNumber:    data.User.PhoneNumber,
				NumberOfGuests: data.User.NumberOfGuests,
			},
			CheckInDate:  data.CheckInDate,
			CheckOutDate: data.CheckOutDate,
			NightlyRates: eventRates,
			TotalAmount:  result.TotalAmount,
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		},
	}

	producerCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.producer.PublishReservationResult(producerCtx, successMsg); err != nil {
		log.Printf("Failed to publish success result for booking %s: %v\n", data.BookingID, err)
	}
}

// publishFailure sends a failure ReservationResultMsg so the booking-service can update the booking status.
func (c *ReservationRequestConsumer) publishFailure(ctx context.Context, msg *event.EventEnvelope[event.BookingCreatedMsg], errorCode string, reason string) {
	data := msg.Data
	failureMsg := event.EventEnvelope[event.ReservationResultMsg]{
		TraceID:   msg.TraceID,
		EventType: "reservation_result",
		Producer:  "room-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: event.ReservationResultMsg{
			BookingID: data.BookingID,
			Success:   false,
			ErrorCode: errorCode,
			Reason:    reason,
			User: event.User{
				UserID:         data.User.UserID,
				Name:           data.User.Name,
				Email:          data.User.Email,
				PhoneNumber:    data.User.PhoneNumber,
				NumberOfGuests: data.User.NumberOfGuests,
			},
			CheckInDate:  data.CheckInDate,
			CheckOutDate: data.CheckOutDate,
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		},
	}

	producerCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.producer.PublishReservationResult(producerCtx, failureMsg); err != nil {
		log.Printf("Failed to publish failure result for booking %s: %v\n", data.BookingID, err)
	}
}
