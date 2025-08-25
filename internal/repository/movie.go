package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MovieReviewRepository struct {
	coll *mongo.Collection
}

func NewMovieReviewRepository(c *mongo.Collection) ReviewRepository {
	return &MovieReviewRepository{
		coll: c,
	}
}

func (r *MovieReviewRepository) GetReviews(ctx context.Context, ID string) ([]Review, error) {
	var result map[string][]Review

	filter := bson.M{"_id": ID}
	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "reviews": 1})
	err := r.coll.FindOne(ctx, filter, opts).Decode(&result)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return []Review{}, ErrNotFound
	} else if err != nil {
		return []Review{}, fmt.Errorf("failed to find review for userID %v: %w", ID, err)
	}
	return result["reviews"], nil
}

func (r *MovieReviewRepository) CreateReview(ctx context.Context, review Review) error {
	filter := bson.M{"_id": review.MovieID, "reviews.userID": review.UserID}
	mResult := r.coll.FindOne(ctx, filter)

	if err := mResult.Err(); err == nil {
		fmt.Printf("err: %v", err)
		return ErrAlreadyExists
	}

	opts := options.UpdateOne().SetUpsert(true)

	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": review.MovieID}, bson.M{
		"$push": map[string]interface{}{
			"reviews": map[string]string{
				"userID": review.UserID,
				"text":   review.Text,
			},
		},
	},
		opts,
	)

	if err != nil {
		return fmt.Errorf("failed to insert review %v: %w", review, err)
	}

	return nil
}

func (r *MovieReviewRepository) UpdateReview(ctx context.Context, review Review) error {
	filter := bson.M{"_id": review.MovieID, "reviews.userID": review.UserID}

	userUpdResult := r.coll.FindOneAndUpdate(ctx, filter, bson.M{
		"$set": map[string]string{
			"reviews.$.text": review.Text,
		},
	})

	if err := userUpdResult.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to update %v: %w", review, err)
	}

	return nil
}
