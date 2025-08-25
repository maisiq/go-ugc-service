package app

import (
	"context"
	"errors"
	"io/fs"

	"github.com/maisiq/go-ugc-service/internal/cache"
	"github.com/maisiq/go-ugc-service/internal/closer"
	"github.com/maisiq/go-ugc-service/internal/db"
	"github.com/maisiq/go-ugc-service/internal/handler"
	"github.com/maisiq/go-ugc-service/internal/producer"
	"github.com/maisiq/go-ugc-service/internal/repository"
	"github.com/maisiq/go-ugc-service/internal/service"
	"github.com/maisiq/go-ugc-service/pkg/config"
	logx "github.com/maisiq/go-ugc-service/pkg/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

type serviceProvider struct {
	cfg        *config.Config
	userRepo   repository.ReviewRepository
	movieRepo  repository.ReviewRepository
	cacher     cache.Cache
	dbConnPool *mongo.Client
	service    *service.UGCService
	broker     *producer.KafkaProducer
	ugcImpl    *handler.UGCServiceServer
	log        *zap.SugaredLogger
	uow        db.UOW
}

func newServiceProvider(cfg *config.Config) *serviceProvider {
	return &serviceProvider{
		cfg: cfg,
	}
}

func (s *serviceProvider) Logger() *zap.SugaredLogger {
	if s.log == nil {
		log := logx.InitLogger(s.cfg.App.Debug)
		s.log = log

		closer.Add(func() error {
			var pathErr *fs.PathError
			if err := log.Sync(); err != nil && !errors.As(err, &pathErr) {
				return err
			}
			return nil
		})
	}
	return s.log
}

func (s *serviceProvider) DBConnPool(ctx context.Context) *mongo.Client {
	if s.dbConnPool == nil {
		pool := db.GetMongoClient(s.cfg.Database, s.Logger())
		s.dbConnPool = pool

		closer.Add(func() error {
			s.Logger().Info("Discard db pool")
			pool.Disconnect(ctx)
			return nil
		})
	}
	return s.dbConnPool
}

func (s *serviceProvider) getUserRepo(ctx context.Context) repository.ReviewRepository {
	if s.userRepo == nil {
		dbName := s.cfg.Database.Name
		collName := s.cfg.Database.Collections.Users
		collection := s.DBConnPool(ctx).Database(dbName).Collection(collName)
		s.userRepo = repository.NewUserReviewRepository(collection)
	}
	return s.userRepo
}

func (s *serviceProvider) UOW(ctx context.Context) db.UOW {
	if s.uow == nil {
		s.uow = db.NewMongoUOW(s.DBConnPool(ctx))
	}
	return s.uow
}

func (s *serviceProvider) getMovieRepo(ctx context.Context) repository.ReviewRepository {
	if s.movieRepo == nil {
		dbName := s.cfg.Database.Name
		collName := s.cfg.Database.Collections.Movies
		collection := s.DBConnPool(ctx).Database(dbName).Collection(collName)
		s.movieRepo = repository.NewMovieReviewRepository(collection)
	}
	return s.movieRepo
}

func (s *serviceProvider) Producer() *producer.KafkaProducer {
	if s.broker == nil {
		s.broker = producer.New(s.cfg.Kafka, s.Logger())

		closer.Add(func() error {
			s.Logger().Info("Closing kafka writer")
			err := s.broker.Writer.Close()

			if err != nil {
				return err
			}
			return nil
		})
	}
	return s.broker
}

func (s *serviceProvider) Cache() *cache.Cache {
	if s.cacher == (cache.Cache{}) {
		s.cacher = cache.Cache{Client: cache.NewClient(&s.cfg.Cache)}

		closer.Add(func() error {
			s.Logger().Info("Closing cache client")
			err := s.cacher.Client.Close()

			if err != nil {
				return err
			}
			return nil
		})
	}
	return &s.cacher
}

func (s *serviceProvider) Service(ctx context.Context) *service.UGCService {
	if s.service == nil {
		s.service = service.NewUGCService(
			s.getUserRepo(ctx), s.getMovieRepo(ctx), s.Logger(), s.Producer(), s.Cache(), s.UOW(ctx),
		)
	}
	return s.service
}

func (s *serviceProvider) UGCServiceServer(ctx context.Context) *handler.UGCServiceServer {
	if s.ugcImpl == nil {
		s.ugcImpl = handler.NewServer(s.Service(ctx))
	}
	return s.ugcImpl
}
