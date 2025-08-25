package db

import (
	"time"

	"github.com/maisiq/go-ugc-service/pkg/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

func GetMongoClient(cfg config.DatabaseConfig, log *zap.SugaredLogger) *mongo.Client {
	log.Info("Trying to connect to MongoDB")
	client, err := mongo.Connect(
		options.Client().ApplyURI(cfg.DSN),
		options.Client().SetMaxPoolSize(100),
		options.Client().SetServerSelectionTimeout(2*time.Second),
	)
	if err != nil {
		log.Fatalf("failed to create mongo client: %v", err)
	}
	return client
}
