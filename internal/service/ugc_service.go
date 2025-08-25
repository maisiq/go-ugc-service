package service

import (
	"context"
	"errors"
	"time"

	"github.com/maisiq/go-ugc-service/internal/cache"
	"github.com/maisiq/go-ugc-service/internal/db"
	apperrors "github.com/maisiq/go-ugc-service/internal/errors"
	"github.com/maisiq/go-ugc-service/internal/producer"
	"github.com/maisiq/go-ugc-service/internal/repository"
	"go.uber.org/zap"
)

type UGCService struct {
	userRepo  repository.ReviewRepository
	movieRepo repository.ReviewRepository
	log       *zap.SugaredLogger
	producer  producer.Producer
	cache     *cache.Cache
	uow       db.UOW
}

func NewUGCService(
	userRepo repository.ReviewRepository,
	movieRepo repository.ReviewRepository,
	log *zap.SugaredLogger,
	producer producer.Producer,
	cache *cache.Cache,
	uow db.UOW,
) *UGCService {
	return &UGCService{
		userRepo:  userRepo,
		movieRepo: movieRepo,
		log:       log,
		producer:  producer,
		cache:     cache,
		uow:       uow,
	}
}

func (s *UGCService) GetReviews(ctx context.Context, UserID, MovieID string) ([]repository.Review, error) {
	key := cache.BuildKey("review", MovieID, UserID)

	var fn func() ([]repository.Review, error)

	if MovieID == "" {
		fn = func() ([]repository.Review, error) {
			return s.userRepo.GetReviews(ctx, UserID)
		}
	} else if UserID == "" {
		fn = func() ([]repository.Review, error) {
			return s.movieRepo.GetReviews(ctx, MovieID)
		}
	}

	reviews, err := cache.GetOrSet(s.cache, ctx, key, time.Minute, fn)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return []repository.Review{}, apperrors.ErrNotFound
		}
		s.log.Errorw("failed to get review",
			"err", err,
		)
		return []repository.Review{}, apperrors.ErrInternal
	}

	return reviews, nil
}

func (s *UGCService) CreateReview(ctx context.Context, UserID, MovieID, Text string) error {

	review := repository.Review{
		UserID:  UserID,
		MovieID: MovieID,
		Text:    Text,
	}

	err := s.uow.RunWithinTx(ctx, func(ctx context.Context) error {
		err := s.userRepo.CreateReview(ctx, review)
		if err != nil {
			return err
		}

		return s.movieRepo.CreateReview(ctx, review)
	})

	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return apperrors.ErrAlreadyExists
		}
		s.log.Errorf("failed to create review: %v", err)
		return apperrors.ErrInternal
	}

	detachedCtx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))

	go s.producer.WriteMessages(detachedCtx, cancel, []producer.AnalyticsMessage{
		{UserID: review.UserID, MovieID: review.MovieID, TimestampMS: time.Now().Unix()},
	})

	return nil
}

func (s *UGCService) UpdateReview(ctx context.Context, UserID, MovieID, Text string) error {
	review := repository.Review{
		UserID:  UserID,
		MovieID: MovieID,
		Text:    Text,
	}
	err := s.uow.RunWithinTx(ctx, func(ctx context.Context) error {
		err := s.userRepo.UpdateReview(ctx, review)
		if err != nil {
			return err
		}

		return s.movieRepo.UpdateReview(ctx, review)
	})

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperrors.ErrNotFound
		}
		s.log.Errorf("Failed to update review: %v", err)
		return apperrors.ErrInternal
	}

	return nil
}
