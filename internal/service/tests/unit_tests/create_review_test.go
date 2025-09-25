package unit_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	apperrors "github.com/maisiq/go-ugc-service/internal/errors"
	"github.com/maisiq/go-ugc-service/internal/producer"
	prodMocks "github.com/maisiq/go-ugc-service/internal/producer/mocks"
	"github.com/maisiq/go-ugc-service/internal/repository"
	repoMocks "github.com/maisiq/go-ugc-service/internal/repository/mocks"
	"github.com/maisiq/go-ugc-service/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCreateReview(t *testing.T) {
	t.Parallel()
	var (
		userID     = gofakeit.UUID()
		movieID    = gofakeit.UUID()
		reviewText = gofakeit.Comment()
		ctx        = context.Background()
		_          = []producer.AnalyticsMessage{
			{UserID: userID, MovieID: movieID, TimestampMS: time.Now().Unix()},
		}
	)

	t.Run("Create review returns no error", func(t *testing.T) {
		t.Parallel()
		uowMocked := repoMocks.NewUOWMock(t)
		producerMocked := prodMocks.NewProducerMock(t)
		s := service.NewUGCService(nil, nil, nil, producerMocked, nil, uowMocked)
		done := make(chan struct{})

		uowMocked.RunWithinTxMock.Return(nil)
		producerMocked.WriteMessagesMock.Set(func(ctx context.Context, cancel context.CancelFunc, messages []producer.AnalyticsMessage) {
			close(done)
		})

		err := s.CreateReview(ctx, userID, movieID, reviewText)
		require.NoError(t, err)

		<-done

	})

	t.Run("Create review returns ErrAlreadyExists", func(t *testing.T) {
		t.Parallel()
		uowMocked := repoMocks.NewUOWMock(t)
		producerMocked := prodMocks.NewProducerMock(t)
		s := service.NewUGCService(nil, nil, nil, producerMocked, nil, uowMocked)

		uowMocked.RunWithinTxMock.Return(repository.ErrAlreadyExists)

		err := s.CreateReview(ctx, userID, movieID, reviewText)
		require.ErrorIs(t, err, apperrors.ErrAlreadyExists)

	})

	t.Run("Create review method writes message to the broker", func(t *testing.T) {
		t.Parallel()
		uowMocked := repoMocks.NewUOWMock(t)
		producerMocked := prodMocks.NewProducerMock(t)
		s := service.NewUGCService(nil, nil, nil, producerMocked, nil, uowMocked)
		done := make(chan struct{})

		uowMocked.RunWithinTxMock.Return(nil)
		producerMocked.WriteMessagesMock.Set(func(ctx context.Context, cancel context.CancelFunc, msgs []producer.AnalyticsMessage) {
			defer close(done)
			if len(msgs) != 1 {
				t.Errorf("expected 1 message, got %d", len(msgs))
				return
			}

			msg := msgs[0]

			if msg.UserID != userID || msg.MovieID != movieID {
				t.Errorf("unexpected user or movie ID: %+v", msg)
			}

			now := time.Now().Unix()
			if msg.TimestampMS < now-2 || msg.TimestampMS > now+2 {
				t.Errorf("timestamp is not recent: %d", msg.TimestampMS)
			}
		})

		err := s.CreateReview(ctx, userID, movieID, reviewText)
		require.NoError(t, err)

		<-done

	})

}
