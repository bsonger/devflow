package service

import (
	"context"
	"github.com/bsonger/devflow/pkg/argo"

	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var JobService = NewJobService()

type jobService struct{}

func NewJobService() *jobService {
	return &jobService{}
}

func (s *jobService) Create(ctx context.Context, job *model.Job) (primitive.ObjectID, error) {
	var err error

	application, err := ApplicationService.Get(ctx, job.ApplicationId)
	if err != nil {
		return primitive.NilObjectID, err
	}
	job.ApplicationName = application.Name
	manifest, err := ManifestService.Get(ctx, job.ApplicationId)
	if err != nil {
		return primitive.NilObjectID, err
	}
	job.ManifestName = manifest.Name

	if job.Type == "install" {
		err = argo.CreateApplication(ctx, job)
	} else {
		err = argo.UpdateApplication(ctx, job)
	}

	if err != nil {
		return primitive.NilObjectID, err
	}
	err = db.Repo.Create(ctx, job)
	if err != nil {
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
