package kafka

import (
	"booking/room-service/event"
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// TODO: Ordering - idempotency - exactly-once semantics
type ReservationProducer struct {
	client *kgo.Client
}

func NewReservationProducer(brokers []string) (*ReservationProducer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()), //Strong consistency: wait for all in-sync replicas to acknowledge
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &ReservationProducer{client: cl}, nil
}

func (p *ReservationProducer) PublishReservationResult(ctx context.Context, msg event.EventEnvelope[event.ReservationResultMsg]) error {
	return p.publish(ctx, RoomReservationEventsTopic, msg)
}

func (p *ReservationProducer) publish(ctx context.Context, topic string, msg any) error {
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
func (p *ReservationProducer) Close() {
	p.client.Close()
}
