package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.uber.org/zap"
)

var Repo *Repository // 全局唯一 Repo

func InitMongo(ctx context.Context, uri, dbName string, logger *zap.Logger) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(uri).
			SetMonitor(otelmongo.NewMonitor()),
	)
	if err != nil {
		return nil, err
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctxPing, nil); err != nil {
		return nil, err
	}

	logger.Info("mongo connected", zap.String("uri", uri))

	Repo = NewRepository(client, dbName, logger) // 全局 repository
	return client, nil
}
