package producer

type AnalyticsMessage struct {
	UserID      string `json:"user_id"`
	MovieID     string `json:"movie_id"`
	TimestampMS int64  `json:"timestamp_ms"`
}
