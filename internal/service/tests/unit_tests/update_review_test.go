package unit_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	apperrors "github.com/maisiq/go-ugc-service/internal/errors"
	"github.com/maisiq/go-ugc-service/internal/repository"
	repoMocks "github.com/maisiq/go-ugc-service/internal/repository/mocks"
	"github.com/maisiq/go-ugc-service/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUpdateReview(t *testing.T) {
	t.Parallel()
	var (
		userID     = gofakeit.UUID()
		movieID    = gofakeit.UUID()
		reviewText = gofakeit.Comment()
		ctx        = context.Background()
	)
	t.Run("Update review returns nil", func(t *testing.T) {
		t.Parallel()

		uowMocked := repoMocks.NewUOWMock(t)
		s := service.NewUGCService(nil, nil, nil, nil, nil, uowMocked)
		uowMocked.RunWithinTxMock.Return(nil)

		err := s.UpdateReview(ctx, userID, movieID, reviewText)

		require.NoError(t, err)

	})

	t.Run("Update review returns ErrNotFound", func(t *testing.T) {
		t.Parallel()

		uowMocked := repoMocks.NewUOWMock(t)
		s := service.NewUGCService(nil, nil, nil, nil, nil, uowMocked)
		uowMocked.RunWithinTxMock.Return(repository.ErrNotFound)

		err := s.UpdateReview(ctx, userID, movieID, reviewText)

		require.ErrorIs(t, err, apperrors.ErrNotFound)

	})
}
