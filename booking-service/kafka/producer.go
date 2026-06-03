package kafka

import (
	"booking/booking-service/event"
	"booking/booking-service/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// TODO: Ordering - idempotency - exactly-once semantics
type BookingProducer struct {
	client *kgo.Client
}

func NewBookingProducer(brokers []string) (*BookingProducer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()), //Strong consistency: wait for all in-sync replicas to acknowledge
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &BookingProducer{client: cl}, nil
}

// PublishBookingCreated builds a lean Kafka message from the domain model
// and publishes it to the booking-request topic.
func (p *BookingProducer) PublishBookingCreated(ctx context.Context, msg event.EventEnvelope[event.BookingCreatedMsg]) error {
	// Resolve trace_id from context, generate if empty
	traceID := logger.GetTraceID(ctx)
	if traceID == "" {
		traceID = uuid.New().String()
		ctx = logger.WithTraceID(ctx, traceID)
	}
	msg.TraceID = traceID

	err := p.publish(ctx, BookingCreatedTopic, msg)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish booking_created event",
			slog.String("trace_id", traceID),
			slog.String("booking_id", msg.Data.BookingID),
			slog.String("topic", BookingCreatedTopic),
			slog.String("error", err.Error()),
		)
		return err
	}

	slog.InfoContext(ctx, "Published booking_created event",
		slog.String("trace_id", traceID),
		slog.String("booking_id", msg.Data.BookingID),
		slog.String("topic", BookingCreatedTopic),
	)
	return nil
}

func (p *BookingProducer) publish(ctx context.Context, topic string, msg any) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	record := &kgo.Record{
		Topic: topic,
		Value: payload,
	}
	return p.client.ProduceSync(ctx, record).FirstErr()
}

// Close gracefully shuts down the producer
func (p *BookingProducer) Close() {
	p.client.Close()
}
