package etl

import (
	"context"
	"time"

	"github.com/maisiq/go-ugc-service/internal/etl/models"
)

func (r *ETLRunner) loader(ctx context.Context, in <-chan models.Msg, batchSize int, flushInterval time.Duration) <-chan models.Msg {
	out := make(chan models.Msg)
	go func() {
		defer close(out)

		batch := make([]models.Msg, 0, batchSize)

		flush := func() {
			if len(batch) == 0 {
				return
			}

			b, err := r.clickhouseConn.PrepareBatch(ctx, "INSERT INTO analytics (user_id, movie_id, timestamp_ms)")

			if err != nil {
				r.log.Errorf("Could not prepare: %v", err)
			}

			for _, res := range batch {
				_ = b.Append(res.Event.UserID, res.Event.MovieID, res.Event.TimestampMS)
			}

			if err := b.Send(); err != nil {
				for _, res := range batch {
					res.Err = err
					out <- res
				}
			}
			batch = batch[:0]

		}

		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				flush()
				return
			case <-ticker.C:
				flush()
			case msg, ok := <-in:
				if !ok {
					flush()
					return
				}
				batch = append(batch, msg)
				if len(batch) >= batchSize {
					flush()
					ticker.Reset(flushInterval)
				}
			}
		}
	}()
	return out
}
