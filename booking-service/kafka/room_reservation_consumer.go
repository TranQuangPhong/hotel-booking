package kafka

import (
	"booking/booking-service/event"
	"booking/booking-service/model"
	"booking/booking-service/pkg/logger"
	"booking/booking-service/service"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
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
			slog.ErrorContext(ctx, "Fetch error",
				slog.String("topic", topic),
				slog.Int("partition", int(partition)),
				slog.String("error", err.Error()),
			)
		})

		// 4. Process the actual messages
		fetches.EachRecord(func(record *kgo.Record) {
			var reservationResult *event.EventEnvelope[event.ReservationResultMsg]

			if err := json.Unmarshal(record.Value, &reservationResult); err != nil {
				// Generate trace_id for unmarshal failures (no envelope available)
				traceID := uuid.New().String()
				msgCtx := logger.WithTraceID(ctx, traceID)
				slog.ErrorContext(msgCtx, "Failed to unmarshal message (skipping)",
					slog.String("trace_id", traceID),
					slog.String("error", err.Error()),
				)
				// Permanent failure — commit to avoid infinite redelivery (poison pill)
				// TODO: publish to dead-letter queue for investigation
				c.client.CommitRecords(ctx, record)
				return
			}

			// Extract trace_id from envelope, generate if empty
			traceID := reservationResult.TraceID
			if traceID == "" {
				traceID = uuid.New().String()
			}
			msgCtx := logger.WithTraceID(ctx, traceID)

			data := reservationResult.Data
			slog.InfoContext(msgCtx, "Processing reservation result",
				slog.String("trace_id", traceID),
				slog.String("booking_id", data.BookingID),
				slog.Bool("success", data.Success),
			)

			// Reservation failed — update status only
			if !data.Success {
				err := c.service.UpdateBookingStatus(msgCtx, data.BookingID, model.StatusReservationFailed)
				if err != nil {
					slog.ErrorContext(msgCtx, "Failed to update booking status to RESERVATION_FAILED",
						slog.String("trace_id", traceID),
						slog.String("booking_id", data.BookingID),
						slog.String("error", err.Error()),
					)
					return // Do NOT commit — message will be redelivered
				}
				slog.InfoContext(msgCtx, "Reservation result processed",
					slog.String("trace_id", traceID),
					slog.String("booking_id", data.BookingID),
					slog.Bool("success", data.Success),
				)
				// DB write succeeded, commit offset
				if err := c.client.CommitRecords(ctx, record); err != nil {
					slog.ErrorContext(msgCtx, "Failed to commit offset",
						slog.String("trace_id", traceID),
						slog.String("booking_id", data.BookingID),
						slog.String("error", err.Error()),
					)
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

			err := c.service.ConfirmBookingReservation(msgCtx, data.BookingID, nightlyRates, data.TotalAmount, data.Room.Price, data.Room.Currency)
			if err != nil {
				slog.ErrorContext(msgCtx, "Failed to confirm booking reservation",
					slog.String("trace_id", traceID),
					slog.String("booking_id", data.BookingID),
					slog.String("error", err.Error()),
				)
				return // Do NOT commit — message will be redelivered
			}
			slog.InfoContext(msgCtx, "Reservation result processed",
				slog.String("trace_id", traceID),
				slog.String("booking_id", data.BookingID),
				slog.Bool("success", data.Success),
			)
			// DB write succeeded, commit offset
			if err := c.client.CommitRecords(ctx, record); err != nil {
				slog.ErrorContext(msgCtx, "Failed to commit offset",
					slog.String("trace_id", traceID),
					slog.String("booking_id", data.BookingID),
					slog.String("error", err.Error()),
				)
			}
		})
	}
}

// Close gracefully shuts down the consumer
func (c *RoomReservationConsumer) Close() {
	c.client.Close()
}
