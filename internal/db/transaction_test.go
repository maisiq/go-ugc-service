package db

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maisiq/go-ugc-service/internal/repository"
	"github.com/maisiq/go-ugc-service/pkg/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUOW(t *testing.T) {
	t.Parallel()
	logx, _ := zap.NewDevelopment()

	var (
		cfg        = config.LoadConfig("../../configs/config.compose.yaml")
		userID     = gofakeit.UUID()
		movieID    = gofakeit.UUID()
		reviewText = gofakeit.Comment()
		review     = repository.Review{UserID: userID, MovieID: movieID, Text: reviewText}
	)

	var (
		ctx              = context.Background()
		log              = logx.Sugar()
		client           = GetMongoClient(cfg.Database, log)
		uow              = NewMongoUOW(client)
		db               = client.Database(cfg.Database.Name)
		usersCollection  = db.Collection("test-user-collection")
		moviesCollection = db.Collection("test-movies-collection")
	)

	CleanDB := func(t *testing.T) {
		t.Cleanup(func() {
			err := db.Drop(ctx)

			if err != nil {
				t.Errorf("Failed to cleanup: %v", err)
			}
		})
	}

	defer CleanDB(t)

	t.Run("User repository returns no results on tx fail", func(t *testing.T) {
		t.Parallel()

		userRepo := repository.NewUserReviewRepository(usersCollection)

		err := uow.RunWithinTx(ctx, func(ctx context.Context) error {
			err := userRepo.CreateReview(ctx, review)
			require.NoError(t, err)

			return repository.ErrAlreadyExists

		})

		require.ErrorIs(t, err, repository.ErrAlreadyExists)
		res, err := userRepo.GetReviews(ctx, review.UserID)

		require.Len(t, res, 0)
		require.ErrorIs(t, err, repository.ErrNotFound)

	})

	t.Run("Movie repository returns no results on tx fail", func(t *testing.T) {
		t.Parallel()

		movieRepo := repository.NewMovieReviewRepository(moviesCollection)

		err := uow.RunWithinTx(ctx, func(ctx context.Context) error {
			err := movieRepo.CreateReview(ctx, review)
			require.NoError(t, err)

			return repository.ErrAlreadyExists

		})

		require.ErrorIs(t, err, repository.ErrAlreadyExists)
		res, err := movieRepo.GetReviews(ctx, review.MovieID)

		require.Len(t, res, 0)
		require.ErrorIs(t, err, repository.ErrNotFound)

	})

	t.Run("Success tx", func(t *testing.T) {
		t.Parallel()

		userRepo := repository.NewUserReviewRepository(usersCollection)
		movieRepo := repository.NewMovieReviewRepository(moviesCollection)

		err := uow.RunWithinTx(ctx, func(ctx context.Context) error {
			err := userRepo.CreateReview(ctx, review)
			require.NoError(t, err)

			err = movieRepo.CreateReview(ctx, review)
			require.NoError(t, err)

			return nil
		})

		require.NoError(t, err)
		res, err := movieRepo.GetReviews(ctx, review.MovieID)

		require.Len(t, res, 1)
		require.NoError(t, err)

		res, err = userRepo.GetReviews(ctx, review.UserID)

		require.Len(t, res, 1)
		require.NoError(t, err)

	})

}
