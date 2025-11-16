package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/model"
)

type ManifestService struct {
}

// NewManifestService 创建 ManifestService
func NewManifestService() *ManifestService {
	return &ManifestService{}
}

// CreateManifest 保存 Manifest 到 Mongo
func (s *ManifestService) CreateManifest(ctx context.Context, m *model.Manifest) (primitive.ObjectID, error) {
	err := db.Repo.Create(ctx, m)
	return m.GetID(), err
}

// GetManifest 根据 ID 查询 Manifest
func (s *ManifestService) GetManifest(ctx context.Context, id primitive.ObjectID) (*model.Manifest, error) {
	m := &model.Manifest{}
	err := db.Repo.FindByID(ctx, m, id)
	return m, err
}

// UpdateManifest 更新 Manifest
func (s *ManifestService) Update(ctx context.Context, m *model.Manifest) error {
	return db.Repo.Update(ctx, m)
}
