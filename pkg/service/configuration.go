package service

import (
	"context"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var ConfigurationService = NewConfigurationService()

type configurationService struct{}

func NewConfigurationService() *configurationService {
	return &configurationService{}
}

func (s *configurationService) Create(ctx context.Context, cfg *model.Configuration) (primitive.ObjectID, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "create_configuration"),
	)

	if err := mongo.Repo.Create(ctx, cfg); err != nil {
		log.Error("create configuration failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log.Info("configuration created", zap.String("configuration_id", cfg.GetID().Hex()))
	return cfg.GetID(), nil
}

func (s *configurationService) Get(ctx context.Context, id primitive.ObjectID) (*model.Configuration, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "get_configuration"),
		zap.String("configuration_id", id.Hex()),
	)

	cfg := &model.Configuration{}
	if err := mongo.Repo.FindByID(ctx, cfg, id); err != nil {
		log.Error("get configuration failed", zap.Error(err))
		return nil, err
	}
	if cfg.DeletedAt != nil {
		log.Warn("configuration already deleted")
		return nil, mongoDriver.ErrNoDocuments
	}

	log.Debug("configuration fetched", zap.String("configuration_name", cfg.Name))
	return cfg, nil
}

func (s *configurationService) Update(ctx context.Context, cfg *model.Configuration) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "update_configuration"),
		zap.String("configuration_id", cfg.GetID().Hex()),
	)

	current := &model.Configuration{}
	if err := mongo.Repo.FindByID(ctx, current, cfg.GetID()); err != nil {
		log.Error("load configuration failed", zap.Error(err))
		return err
	}
	if current.DeletedAt != nil {
		log.Warn("update skipped for deleted configuration")
		return mongoDriver.ErrNoDocuments
	}

	cfg.CreatedAt = current.CreatedAt
	cfg.DeletedAt = current.DeletedAt
	cfg.WithUpdateDefault()

	if err := mongo.Repo.Update(ctx, cfg); err != nil {
		log.Error("update configuration failed", zap.Error(err))
		return err
	}

	log.Debug("configuration updated", zap.String("configuration_name", cfg.Name))
	return nil
}

func (s *configurationService) Delete(ctx context.Context, id primitive.ObjectID) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "delete_configuration"),
		zap.String("configuration_id", id.Hex()),
	)

	now := time.Now()
	update := primitive.M{
		"$set": primitive.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	if err := mongo.Repo.UpdateByID(ctx, &model.Configuration{}, id, update); err != nil {
		log.Error("delete configuration failed", zap.Error(err))
		return err
	}

	log.Info("configuration deleted")
	return nil
}

func (s *configurationService) List(ctx context.Context, filter primitive.M) ([]model.Configuration, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "list_configurations"),
		zap.Any("filter", filter),
	)

	var cfgs []model.Configuration
	if err := mongo.Repo.List(ctx, &model.Configuration{}, filter, &cfgs); err != nil {
		log.Error("list configurations failed", zap.Error(err))
		return nil, err
	}

	log.Debug("configurations listed", zap.Int("count", len(cfgs)))
	return cfgs, nil
}
