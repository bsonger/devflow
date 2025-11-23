package service

import (
	"context"
	"github.com/bsonger/devflow/pkg/argo"
	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var JobService = NewJobService()

type jobService struct{}

func NewJobService() *jobService {
	return &jobService{}
}

func (s *jobService) Create(ctx context.Context, job *model.Job) (primitive.ObjectID, error) {
	var err error

	manifest, err := ManifestService.Get(ctx, job.ManifestID)
	if err != nil {
		logging.Logger.Error("Failed to create job", zap.Error(err))
		return primitive.NilObjectID, err
	}
	job.ManifestName = manifest.Name

	application, err := ApplicationService.Get(ctx, manifest.ApplicationId)
	if err != nil {
		logging.Logger.Error("Failed to create job", zap.Error(err))
		return primitive.NilObjectID, err
	}
	job.ApplicationName = application.Name

	if job.Type == "install" {
		err = argo.CreateApplication(ctx, job)
	} else {
		err = argo.UpdateApplication(ctx, job)
	}

	if err != nil {
		logging.Logger.Error("Failed to create job", zap.Error(err))
		return primitive.NilObjectID, err
	}
	job.WithCreateDefault()
	err = db.Repo.Create(ctx, job)
	if err != nil {
		logging.Logger.Error("Failed to create job", zap.Error(err))
		return primitive.NilObjectID, err
	}
	return job.GetID(), err
}

func (s *jobService) Get(ctx context.Context, id primitive.ObjectID) (*model.Job, error) {
	app := &model.Job{}
	err := db.Repo.FindByID(ctx, app, id)
	return app, err
}

func (s *jobService) Update(ctx context.Context, app *model.Job) error {
	return db.Repo.Update(ctx, app)
}

func (s *jobService) Delete(ctx context.Context, id primitive.ObjectID) error {
	app := &model.Job{}
	return db.Repo.Delete(ctx, app, id)
}

func (s *jobService) List(ctx context.Context, filter primitive.M) ([]*model.Job, error) {
	var apps []*model.Job
	err := db.Repo.List(ctx, &model.Job{}, filter, &apps)
	return apps, err
}
