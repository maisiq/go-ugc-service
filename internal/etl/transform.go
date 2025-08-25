package etl

import (
	"github.com/mailru/easyjson"
	"github.com/maisiq/go-ugc-service/internal/etl/models"
)

func (r *ETLRunner) transform(in <-chan models.Msg) <-chan models.Msg {
	out := make(chan models.Msg)
	go func() {
		defer close(out)
		for msg := range in {
			if msg.Err != nil {
				out <- msg
				continue
			}
			var e models.AnalyticsEvent
			if err := easyjson.Unmarshal(msg.KafkaMsg.Value, &e); err != nil {
				msg.Err = err
			} else {
				msg.Event = e
			}
			out <- msg
		}
	}()
	return out
}
