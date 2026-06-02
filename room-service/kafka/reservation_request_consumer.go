package kafka

import (
	"booking/room-service/event"
	"booking/room-service/model"
	"booking/room-service/pkg/logger"
	"booking/room-service/service"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
			slog.ErrorContext(ctx, "Fetch error",
				slog.String("topic", topic),
				slog.Int("partition", int(partition)),
				slog.String("error", err.Error()),
			)
		})

		// 4. Process the actual messages
		fetches.EachRecord(func(record *kgo.Record) {
			var bookingCreatedMsg *event.EventEnvelope[event.BookingCreatedMsg]

			if err := json.Unmarshal(record.Value, &bookingCreatedMsg); err != nil {
				slog.ErrorContext(ctx, "Failed to unmarshal message (skipping)",
					slog.String("error", err.Error()),
				)
				// Permanent failure — commit to avoid infinite redelivery (poison pill)
				// TODO: publish to dead-letter queue for investigation
				c.client.CommitRecords(ctx, record)
				return
			}

			// Extract TraceID from incoming envelope or generate new one
			traceID := bookingCreatedMsg.TraceID
			if traceID == "" {
				traceID = generateUUID()
			}

			// Create enriched context with trace_id
			msgCtx := logger.WithTraceID(ctx, traceID)

			data := bookingCreatedMsg.Data
			slog.InfoContext(msgCtx, "Processing booking reservation",
				slog.String("booking_id", data.BookingID),
				slog.String("room_id", data.RoomID),
			)

			// Reserve room and calculate pricing from inventory
			result, err := c.service.ReserveRoom(msgCtx, data.RoomID, data.CheckInDate, data.CheckOutDate, data.BookingID)
			if err != nil {
				slog.ErrorContext(msgCtx, "Failed to reserve room",
					slog.String("booking_id", data.BookingID),
					slog.String("room_id", data.RoomID),
					slog.String("error", err.Error()),
				)
				// Publish failure so booking-service can update status
				c.publishFailure(msgCtx, bookingCreatedMsg, traceID, "INTERNAL_ERROR", fmt.Sprintf("failed to reserve room: %v", err))
				return // Do NOT commit — message will be redelivered
			}

			// Fetch room details for the response
			room, err := c.service.GetRoomByID(msgCtx, data.RoomID)
			if err != nil {
				slog.ErrorContext(msgCtx, "Failed to fetch room details",
					slog.String("booking_id", data.BookingID),
					slog.String("room_id", data.RoomID),
					slog.String("error", err.Error()),
				)
				// Publish failure so booking-service can update status
				c.publishFailure(msgCtx, bookingCreatedMsg, traceID, "INTERNAL_ERROR", fmt.Sprintf("failed to fetch room details: %v", err))
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
			c.publishSuccess(msgCtx, bookingCreatedMsg, traceID, result, room, eventRates)
			slog.InfoContext(msgCtx, "Reservation result published",
				slog.String("booking_id", data.BookingID),
				slog.String("room_id", data.RoomID),
				slog.Bool("success", result.Success),
			)

			// DB write succeeded + result published, commit offset
			if err := c.client.CommitRecords(ctx, record); err != nil {
				slog.ErrorContext(msgCtx, "Failed to commit offset",
					slog.String("booking_id", data.BookingID),
					slog.String("error", err.Error()),
				)
			}
		})
	}
}

// Close gracefully shuts down the consumer
func (c *ReservationRequestConsumer) Close() {
	c.client.Close()
}

// publishSuccess sends a success ReservationResultMsg with pricing and room details.
func (c *ReservationRequestConsumer) publishSuccess(ctx context.Context, msg *event.EventEnvelope[event.BookingCreatedMsg], traceID string, result *service.ReservationResult, room *model.Room, eventRates []event.NightlyRate) {
	data := msg.Data
	successMsg := event.EventEnvelope[event.ReservationResultMsg]{
		TraceID:   traceID,
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
		slog.ErrorContext(ctx, "Failed to publish success result",
			slog.String("booking_id", data.BookingID),
			slog.String("topic", RoomReservationEventsTopic),
			slog.String("error", err.Error()),
		)
	}
}

// publishFailure sends a failure ReservationResultMsg so the booking-service can update the booking status.
func (c *ReservationRequestConsumer) publishFailure(ctx context.Context, msg *event.EventEnvelope[event.BookingCreatedMsg], traceID string, errorCode string, reason string) {
	data := msg.Data
	failureMsg := event.EventEnvelope[event.ReservationResultMsg]{
		TraceID:   traceID,
		EventType: "reservation_result",
		Producer:  "room-service",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data: event.ReservationResultMsg{
			BookingID: data.BookingID,
			Success:   false,
			ErrorCode: errorCode,
			Reason:    reason,
			Room: event.Room{
				RoomID: data.RoomID,
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
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		},
	}

	producerCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.producer.PublishReservationResult(producerCtx, failureMsg); err != nil {
		slog.ErrorContext(ctx, "Failed to publish failure result",
			slog.String("booking_id", data.BookingID),
			slog.String("topic", RoomReservationEventsTopic),
			slog.String("error", err.Error()),
		)
	}
}
