package repository

import "context"

//go:generate minimock -i ReviewRepository -o ./mocks/ -s "_mock.go"
type ReviewRepository interface {
	GetReviews(ctx context.Context, ID string) ([]Review, error)
	CreateReview(ctx context.Context, review Review) error
	UpdateReview(ctx context.Context, review Review) error
}
