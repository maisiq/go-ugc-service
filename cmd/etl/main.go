package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/maisiq/go-ugc-service/internal/closer"
	"github.com/maisiq/go-ugc-service/internal/etl"
	"github.com/maisiq/go-ugc-service/internal/etl/clickhouse"
	"github.com/maisiq/go-ugc-service/pkg/config"
	"github.com/maisiq/go-ugc-service/pkg/logger"
	"github.com/segmentio/kafka-go"
)

func main() {
	env, ok := os.LookupEnv("environment")

	if !ok {
		env = "local"
	}
	cfgPath := fmt.Sprintf("./configs/config.%s.yaml", env)

	cfg := config.LoadETLConfig(cfgPath)
	log := logger.InitLogger(cfg.App.Debug)
	ctx := context.Background()
	c := closer.New(os.Interrupt, syscall.SIGTERM)

	ch, _ := clickhouse.InitClickhouseClient(ctx, &cfg.Clickhouse)

	c.Add(func() error {
		log.Debug("Закрываю подключение к clickhouse")
		err := ch.Close()

		if err != nil {
			log.Errorf("Ошибка при закрытие подключения с clickhouse: %v", err)
		}
		return err
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.AnalyticsTopic,
		GroupID: cfg.Consumer.GroupID,
	})

	c.Add(func() error {
		log.Debug("Закрываю подключение к kafka")
		err := reader.Close()

		if err != nil {
			log.Errorf("Ошибка при закрытие подключения с kafka: %v", err)
		}
		return err
	})

	runner := etl.NewRunner(log, ch, cfg, reader)
	go runner.Run(ctx)

	c.Wait()
}
