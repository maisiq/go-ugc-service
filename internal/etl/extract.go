package etl

import (
	"context"

	"github.com/maisiq/go-ugc-service/internal/etl/models"
)

func (r *ETLRunner) readMessages(ctx context.Context) <-chan models.Msg {
	out := make(chan models.Msg)
	go func() {
		defer close(out)
		for {
			m, err := r.kafkaReader.FetchMessage(ctx)
			if err != nil {
				r.log.Errorf("Kafka read error:", err)
				return
			}
			out <- models.Msg{KafkaMsg: m}
		}
	}()
	return out
}
