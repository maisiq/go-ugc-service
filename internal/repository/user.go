package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserReviewRepository struct {
	coll *mongo.Collection
}

func NewUserReviewRepository(c *mongo.Collection) ReviewRepository {
	return &UserReviewRepository{
		coll: c,
	}
}

func (r *UserReviewRepository) GetReviews(ctx context.Context, ID string) ([]Review, error) {
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

func (r *UserReviewRepository) CreateReview(ctx context.Context, review Review) error {
	filter := bson.M{"_id": review.UserID, "reviews.movieID": review.MovieID}
	mResult := r.coll.FindOne(ctx, filter)

	if err := mResult.Err(); err == nil {
		fmt.Printf("err: %v", err)
		return ErrAlreadyExists
	}

	opts := options.UpdateOne().SetUpsert(true)

	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": review.UserID}, bson.M{
		"$push": map[string]interface{}{
			"reviews": map[string]string{
				"movieID": review.MovieID,
				"text":    review.Text,
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

func (r *UserReviewRepository) UpdateReview(ctx context.Context, review Review) error {
	filter := bson.M{"_id": review.UserID, "reviews.movieID": review.MovieID}

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
