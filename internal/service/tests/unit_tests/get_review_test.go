package unit_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/maisiq/go-ugc-service/internal/cache"
	apperrors "github.com/maisiq/go-ugc-service/internal/errors"
	"github.com/maisiq/go-ugc-service/internal/producer"
	"github.com/maisiq/go-ugc-service/internal/repository"
	repoMocks "github.com/maisiq/go-ugc-service/internal/repository/mocks"
	"github.com/maisiq/go-ugc-service/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetReview(t *testing.T) {
	t.Parallel()

	log, _ := zap.NewDevelopment()

	var (
		userID     = gofakeit.UUID()
		movieID    = gofakeit.UUID()
		reviewText = gofakeit.Comment()
		ctx        = context.Background()
		reviewsExp = []repository.Review{
			{UserID: userID, MovieID: movieID, Text: reviewText},
		}
		_ = []producer.AnalyticsMessage{
			{UserID: userID, MovieID: movieID, TimestampMS: time.Now().Unix()},
		}
		sugLogger = log.Sugar()
	)
	t.Run("Get user reviews returns review, no cache", func(t *testing.T) {
		t.Parallel()

		rs := miniredis.RunT(t)
		c := redis.NewClient(&redis.Options{Addr: rs.Addr()})
		cache := &cache.Cache{Client: c}

		repoMocked := repoMocks.NewReviewRepositoryMock(t)
		s := service.NewUGCService(repoMocked, repoMocked, nil, nil, cache, nil)

		repoMocked.GetReviewsMock.Expect(ctx, userID).Return(reviewsExp, nil)
		review, err := s.GetReviews(ctx, userID, "")

		require.NoError(t, err)
		require.Equal(t, reviewsExp, review)

	})

	t.Run("Get user reviews returns review using cache", func(t *testing.T) {
		t.Parallel()

		rs := miniredis.RunT(t)
		c := redis.NewClient(&redis.Options{Addr: rs.Addr()})
		cache := &cache.Cache{Client: c}

		key := fmt.Sprintf("cache:%v:%v", "review", userID)
		b, _ := json.Marshal(reviewsExp)
		rs.Set(key, string(b))

		repoMocked := repoMocks.NewReviewRepositoryMock(t)
		s := service.NewUGCService(repoMocked, repoMocked, nil, nil, cache, nil)

		review, err := s.GetReviews(ctx, userID, "")

		require.NoError(t, err)
		require.Equal(t, reviewsExp, review)

	})

	t.Run("Get user reviews returns ErrNotFound", func(t *testing.T) {
		t.Parallel()

		rs := miniredis.RunT(t)
		c := redis.NewClient(&redis.Options{Addr: rs.Addr()})
		cache := &cache.Cache{Client: c}

		repoMocked := repoMocks.NewReviewRepositoryMock(t)
		repoMocked.GetReviewsMock.Return([]repository.Review{}, repository.ErrNotFound)
		s := service.NewUGCService(repoMocked, repoMocked, sugLogger, nil, cache, nil)

		review, err := s.GetReviews(ctx, userID, "")

		require.ErrorIs(t, err, apperrors.ErrNotFound)
		require.Equal(t, []repository.Review{}, review)

	})
}
