package kafka

import (
	"booking/booking-service/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

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

func (p *BookingProducer) PublishCreateBookingMessage(ctx context.Context, booking *model.Booking) error {
	return p.publish(ctx, BookingRequestTopic, booking)
}

func (p *BookingProducer) publish(ctx context.Context, topic string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	record := &kgo.Record{
		Topic: topic,
		Value: data,
	}
	return p.client.ProduceSync(ctx, record).FirstErr()
}
