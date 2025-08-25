package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/maisiq/go-ugc-service/pkg/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

//go:generate minimock -i Producer -o mocks/producer_mock.go
type Producer interface {
	WriteMessages(ctx context.Context, cancel context.CancelFunc, messages []AnalyticsMessage)
}

type KafkaProducer struct {
	Writer *kafka.Writer
	log    *zap.SugaredLogger
}

func New(cfg config.KafkaConfig, log *zap.SugaredLogger) *KafkaProducer {
	log.Debug("Initializing new Kafka Writer")

	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.AnalyticsTopic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
	}

	return &KafkaProducer{Writer: w, log: log}
}

func (p *KafkaProducer) WriteMessages(ctx context.Context, cancel context.CancelFunc, messages []AnalyticsMessage) {
	defer cancel()
	var KafkaMessages []kafka.Message

	for _, msg := range messages {
		rawMsg, err := json.Marshal(msg)

		if err != nil {
			p.log.Errorf("Failed to parse message: %+v", msg)
		}

		KafkaMessages = append(KafkaMessages, kafka.Message{Value: rawMsg})
	}

	err := p.Writer.WriteMessages(
		ctx,
		KafkaMessages...,
	)

	if err != nil {
		p.log.Errorf("failed to write messages:", err)
	}

	p.log.Debug("Wrote to the broker")

}
