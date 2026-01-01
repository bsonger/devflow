package service

import (
	"context"

	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ConfigurationService = NewConfigurationService()

type configurationService struct{}

func NewConfigurationService() *configurationService {
	return &configurationService{}
}

func (s *configurationService) Create(ctx context.Context, cfg *model.Configuration) (primitive.ObjectID, error) {
	err := mongo.Repo.Create(ctx, cfg)
	return cfg.GetID(), err
}

func (s *configurationService) Get(ctx context.Context, id primitive.ObjectID) (*model.Configuration, error) {
	cfg := &model.Configuration{}
	err := mongo.Repo.FindByID(ctx, cfg, id)
	return cfg, err
}

func (s *configurationService) Update(ctx context.Context, cfg *model.Configuration) error {
	return mongo.Repo.Update(ctx, cfg)
}

func (s *configurationService) Delete(ctx context.Context, id primitive.ObjectID) error {
	cfg := &model.Configuration{}
	return mongo.Repo.Delete(ctx, cfg, id)
}

func (s *configurationService) List(ctx context.Context, filter primitive.M) ([]model.Configuration, error) {
	var cfgs []model.Configuration
	err := mongo.Repo.List(ctx, &model.Configuration{}, filter, &cfgs)
	return cfgs, err
}
