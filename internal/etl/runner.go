package etl

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/maisiq/go-ugc-service/internal/etl/models"
	"github.com/maisiq/go-ugc-service/pkg/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type ETLRunner struct {
	log            *zap.SugaredLogger
	clickhouseConn driver.Conn
	cfg            *config.ETLConfig
	kafkaReader    *kafka.Reader
}

func NewRunner(log *zap.SugaredLogger, clickhouseConn driver.Conn, cfg *config.ETLConfig, kafkaReader *kafka.Reader) *ETLRunner {
	return &ETLRunner{
		log:            log,
		clickhouseConn: clickhouseConn,
		cfg:            cfg,
		kafkaReader:    kafkaReader,
	}
}

func (r *ETLRunner) Run(ctx context.Context) {
	raw := r.readMessages(ctx)
	parsed := r.transform(raw)
	saved := r.loader(ctx, parsed, 5, 5*time.Minute)
	r.commit(ctx, saved)
}

func (r *ETLRunner) commit(ctx context.Context, in <-chan models.Msg) {
	for res := range in {
		if res.Err != nil {
			r.log.Errorf("Proccess error:", res.Err)
			continue
		}
		if err := r.kafkaReader.CommitMessages(ctx, res.KafkaMsg); err != nil {
			r.log.Errorf("Kafka commit error:", err)
		}
	}
}
