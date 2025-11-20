package db

import (
	"context"

	"github.com/bsonger/devflow/pkg/model"
	"github.com/bsonger/devflow/pkg/otel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Repository struct {
	client *mongo.Client
	dbName string
	logger *zap.Logger
}

func NewRepository(client *mongo.Client, dbName string, logger *zap.Logger) *Repository {
	return &Repository{
		client: client,
		dbName: dbName,
		logger: logger,
	}
}

func (r *Repository) collection(m model.MongoModel) *mongo.Collection {
	return r.client.Database(r.dbName).Collection(m.CollectionName())
}

func (r *Repository) Create(ctx context.Context, m model.MongoModel) error {
	ctx, span := otel.Start(ctx, "repo.create")
	defer span.End()

	if m.GetID().IsZero() {
		m.SetID(primitive.NewObjectID())
	}

	_, err := r.collection(m).InsertOne(ctx, m)
	return err
}

func (r *Repository) FindByID(ctx context.Context, m model.MongoModel, id primitive.ObjectID) error {
	ctx, span := otel.Start(ctx, "repo.findById")
	defer span.End()

	return r.collection(m).FindOne(ctx, bson.M{"_id": id}).Decode(m)
}

func (r *Repository) Update(ctx context.Context, m model.MongoModel) error {
	ctx, span := otel.Start(ctx, "repo.update")
	defer span.End()

	_, err := r.collection(m).
		UpdateByID(ctx, m.GetID(), bson.M{"$set": m})

	return err
}

func (r *Repository) Delete(ctx context.Context, m model.MongoModel, id primitive.ObjectID) error {
	_, err := r.collection(m).
		UpdateByID(ctx, id, bson.M{"$set": bson.M{"deleted": true}})
	return err
}

func (r *Repository) List(ctx context.Context, m model.MongoModel, filter bson.M, results interface{}) error {
	ctx, span := otel.Start(ctx, "repo.list")
	defer span.End()

	if filter == nil {
		filter = bson.M{}
	}

	cur, err := r.collection(m).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	// cur.All 会把所有文档解码到 results（results 必须是 slice 的指针）
	if err := cur.All(ctx, results); err != nil {
		return err
	}
	return nil
}
