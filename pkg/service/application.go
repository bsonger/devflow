package service

import (
	"context"

	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApplicationService struct{}

func NewApplicationService() *ApplicationService {
	return &ApplicationService{}
}

func (s *ApplicationService) Create(ctx context.Context, app *model.Application) (primitive.ObjectID, error) {
	err := db.Repo.Create(ctx, app)
	return app.GetID(), err
}

func (s *ApplicationService) Get(ctx context.Context, id primitive.ObjectID) (*model.Application, error) {
	app := &model.Application{}
	err := db.Repo.FindByID(ctx, app, id)
	return app, err
}

func (s *ApplicationService) Update(ctx context.Context, app *model.Application) error {
	return db.Repo.Update(ctx, app)
}

func (s *ApplicationService) Delete(ctx context.Context, id primitive.ObjectID) error {
	app := &model.Application{}
	return db.Repo.Delete(ctx, app, id)
}

func (s *ApplicationService) List(ctx context.Context, filter primitive.M) ([]model.Application, error) {
	var apps []model.Application
	err := db.Repo.List(ctx, &model.Application{}, filter, &apps)
	return apps, err
}
