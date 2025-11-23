package service

import (
	"context"
	"errors"
	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/model"
	"github.com/bsonger/devflow/pkg/tekton"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ManifestService = NewManifestService()

type manifestService struct {
}

// NewManifestService 创建 ManifestService
func NewManifestService() *manifestService {
	return &manifestService{}
}

// CreateManifest 保存 Manifest 到 Mongo
func (s *manifestService) CreateManifest(ctx context.Context, m *model.Manifest) (primitive.ObjectID, error) {

	application, err := ApplicationService.Get(ctx, m.ApplicationId)
	if err != nil {
		return primitive.NilObjectID, errors.New("application is not found")
	}

	m.GitRepo = application.RepoURL

	// 自动生成 Manifest 名称
	m.ApplicationName = application.Name

	m.Name = model.GenerateManifestVersion(m.ApplicationName)
	m.WithCreateDefault()

	pipelineRun, err := tekton.CreatePipelineRun(ctx, "devflow-ci", m)
	if err != nil {
		return primitive.NilObjectID, err
	}
	m.PipelineID = pipelineRun.Name
	err = db.Repo.Create(ctx, m)
	return m.GetID(), err
}

// GetManifest 根据 ID 查询 Manifest
func (s *manifestService) GetManifest(ctx context.Context, id primitive.ObjectID) (*model.Manifest, error) {
	m := &model.Manifest{}
	err := db.Repo.FindByID(ctx, m, id)
	return m, err
}

// UpdateManifest 更新 Manifest
func (s *manifestService) Update(ctx context.Context, m *model.Manifest) error {
	return db.Repo.Update(ctx, m)
}

func (s *manifestService) List(ctx context.Context, filter primitive.M) ([]model.Manifest, error) {
	var apps []model.Manifest
	err := db.Repo.List(ctx, &model.Manifest{}, filter, &apps)
	return apps, err
}

func (s *manifestService) Get(ctx context.Context, id primitive.ObjectID) (*model.Manifest, error) {
	app := &model.Manifest{}
	err := db.Repo.FindByID(ctx, app, id)
	return app, err
}
