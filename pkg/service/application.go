package service

import (
	"context"
	"errors"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var ApplicationService = NewApplicationService()

var ErrManifestNotForApplication = errors.New("manifest does not belong to application")

type applicationService struct{}

func NewApplicationService() *applicationService {
	return &applicationService{}
}

// Create 创建 Application
func (s *applicationService) Create(ctx context.Context, app *model.Application) (primitive.ObjectID, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "create_application"),
	)

	if err := mongo.Repo.Create(ctx, app); err != nil {
		log.Error("create application failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log.Info("application created", zap.String("application_id", app.GetID().Hex()))
	return app.GetID(), nil
}

// Get 根据 ID 查询 Application
func (s *applicationService) Get(ctx context.Context, id primitive.ObjectID) (*model.Application, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "get_application"),
		zap.String("application_id", id.Hex()),
	)

	app := &model.Application{}
	if err := mongo.Repo.FindByID(ctx, app, id); err != nil {
		log.Error("get application failed", zap.Error(err))
		return nil, err
	}
	if app.DeletedAt != nil {
		log.Warn("application already deleted")
		return nil, mongoDriver.ErrNoDocuments
	}

	log.Debug("application fetched", zap.String("application_name", app.Name))
	return app, nil
}

// Update 更新 Application
func (s *applicationService) Update(ctx context.Context, app *model.Application) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "update_application"),
		zap.String("application_id", app.GetID().Hex()),
	)

	current := &model.Application{}
	if err := mongo.Repo.FindByID(ctx, current, app.GetID()); err != nil {
		log.Error("load application failed", zap.Error(err))
		return err
	}
	if current.DeletedAt != nil {
		log.Warn("update skipped for deleted application")
		return mongoDriver.ErrNoDocuments
	}

	app.CreatedAt = current.CreatedAt
	app.DeletedAt = current.DeletedAt
	app.WithUpdateDefault()

	if err := mongo.Repo.Update(ctx, app); err != nil {
		log.Error("update application failed", zap.Error(err))
		return err
	}

	log.Debug("application updated", zap.String("application_name", app.Name))
	return nil
}

// Delete 删除 Application
func (s *applicationService) Delete(ctx context.Context, id primitive.ObjectID) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "delete_application"),
		zap.String("application_id", id.Hex()),
	)

	now := time.Now()
	update := primitive.M{
		"$set": primitive.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	if err := mongo.Repo.UpdateByID(ctx, &model.Application{}, id, update); err != nil {
		log.Error("delete application failed", zap.Error(err))
		return err
	}

	log.Info("application deleted")
	return nil
}

// UpdateActiveManifest updates the application active manifest reference.
func (s *applicationService) UpdateActiveManifest(ctx context.Context, appID, manifestID primitive.ObjectID) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "update_application_active_manifest"),
		zap.String("application_id", appID.Hex()),
		zap.String("manifest_id", manifestID.Hex()),
	)

	app := &model.Application{}
	if err := mongo.Repo.FindByID(ctx, app, appID); err != nil {
		log.Error("get application failed", zap.Error(err))
		return err
	}
	if app.DeletedAt != nil {
		log.Warn("application already deleted")
		return mongoDriver.ErrNoDocuments
	}

	manifest := &model.Manifest{}
	if err := mongo.Repo.FindByID(ctx, manifest, manifestID); err != nil {
		log.Error("get manifest failed", zap.Error(err))
		return err
	}
	if manifest.DeletedAt != nil {
		log.Warn("manifest already deleted")
		return mongoDriver.ErrNoDocuments
	}
	if manifest.ApplicationId != appID {
		log.Warn("manifest does not belong to application")
		return ErrManifestNotForApplication
	}

	update := primitive.M{
		"$set": primitive.M{
			"active_manifest_id":   manifestID,
			"active_manifest_name": manifest.Name,
			"updated_at":           time.Now(),
		},
	}

	if err := mongo.Repo.UpdateByID(ctx, &model.Application{}, appID, update); err != nil {
		log.Error("update active manifest failed", zap.Error(err))
		return err
	}

	log.Info("active manifest updated", zap.String("active_manifest_name", manifest.Name))
	return nil
}

// List 查询 Application 列表
func (s *applicationService) List(ctx context.Context, filter primitive.M) ([]model.Application, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "list_applications"),
		zap.Any("filter", filter),
	)

	var apps []model.Application
	if err := mongo.Repo.List(ctx, &model.Application{}, filter, &apps); err != nil {
		log.Error("list applications failed", zap.Error(err))
		return nil, err
	}

	log.Debug("applications listed", zap.Int("count", len(apps)))
	return apps, nil
}
