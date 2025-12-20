package service

import (
	"context"

	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ApplicationService = NewApplicationService()

type applicationService struct{}

func NewApplicationService() *applicationService {
	return &applicationService{}
}

func (s *applicationService) Create(ctx context.Context, app *model.Application) (primitive.ObjectID, error) {
	err := mongo.Repo.Create(ctx, app)
	return app.GetID(), err
}

func (s *applicationService) Get(ctx context.Context, id primitive.ObjectID) (*model.Application, error) {
	app := &model.Application{}
	err := mongo.Repo.FindByID(ctx, app, id)
	return app, err
}

func (s *applicationService) Update(ctx context.Context, app *model.Application) error {
	return mongo.Repo.Update(ctx, app)
}

func (s *applicationService) Delete(ctx context.Context, id primitive.ObjectID) error {
	app := &model.Application{}
	return mongo.Repo.Delete(ctx, app, id)
}

func (s *applicationService) List(ctx context.Context, filter primitive.M) ([]model.Application, error) {
	var apps []model.Application
	err := mongo.Repo.List(ctx, &model.Application{}, filter, &apps)
	return apps, err
}
