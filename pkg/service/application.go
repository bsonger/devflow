package service

import (
	"context"
	"go.uber.org/zap"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ApplicationService = NewApplicationService()

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

	log.Debug("application fetched", zap.String("application_name", app.Name))
	return app, nil
}

// Update 更新 Application
func (s *applicationService) Update(ctx context.Context, app *model.Application) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "update_application"),
		zap.String("application_id", app.GetID().Hex()),
	)

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

	app := &model.Application{}
	if err := mongo.Repo.Delete(ctx, app, id); err != nil {
		log.Error("delete application failed", zap.Error(err))
		return err
	}

	log.Info("application deleted")
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
