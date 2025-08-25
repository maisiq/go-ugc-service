package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

//go:generate minimock -i UOW -o ./mocks/ -s "_mock.go"
type UOW interface {
	RunWithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type MongoUOW struct {
	client *mongo.Client
}

func NewMongoUOW(client *mongo.Client) UOW {
	return &MongoUOW{
		client: client,
	}
}

func (uow *MongoUOW) RunWithinTx(ctx context.Context, fn func(ctx context.Context) error) error {

	sess, err := uow.client.StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)

	// sess.WithTransaction()
	err = mongo.WithSession(ctx, sess, func(ctx context.Context) error {
		if err := sess.StartTransaction(); err != nil {
			return err
		}

		err := fn(ctx)

		if err != nil {
			sess.AbortTransaction(ctx)
			return err
		}

		if err := sess.CommitTransaction(ctx); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
