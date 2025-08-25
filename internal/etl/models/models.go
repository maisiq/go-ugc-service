package models

import "github.com/segmentio/kafka-go"

//easyjson:json
type AnalyticsEvent struct {
	UserID      string `json:"user_id"`
	MovieID     string `json:"movie_id"`
	TimestampMS int64  `json:"timestamp_ms"`
}

type Msg struct {
	KafkaMsg kafka.Message
	Event    AnalyticsEvent
	Err      error
}
