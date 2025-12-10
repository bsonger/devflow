package service

import (
	"context"
	"github.com/bsonger/devflow/pkg/argo"
	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var JobService = NewJobService()

type jobService struct{}

func NewJobService() *jobService {
	return &jobService{}
}

func (s *jobService) Create(ctx context.Context, job *model.Job) (primitive.ObjectID, error) {
	tracer := otel.Tracer("devflow-job")
	ctx, span := tracer.Start(ctx, "CreateJob")
	defer span.End()

	// 从 Trace Context 自动注入 trace_id 到日志
	logger := logging.Logger.With(zap.String("trace_id", trace.SpanFromContext(ctx).SpanContext().TraceID().String()))

	// 获取 Manifest
	manifest, err := ManifestService.Get(ctx, job.ManifestID)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to get manifest", zap.Error(err))
		return primitive.NilObjectID, err
	}
	job.ManifestName = manifest.Name
	job.ApplicationId = manifest.ApplicationId

	// 获取 Application
	application, err := ApplicationService.Get(ctx, manifest.ApplicationId)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to get application", zap.Error(err))
		return primitive.NilObjectID, err
	}
	job.ApplicationName = application.Name

	// 调用 Argo 创建/更新 Application
	var argoSpan trace.Span
	ctx, argoSpan = tracer.Start(ctx, "ArgoCreateOrUpdate")
	if job.Type == "install" {
		err = argo.CreateApplication(ctx, job)
	} else {
		err = argo.UpdateApplication(ctx, job)
	}
	if err != nil {
		argoSpan.RecordError(err)
		logger.Error("Failed to create/update application", zap.Error(err))
		argoSpan.End()
		span.RecordError(err)
		return primitive.NilObjectID, err
	}
	argoSpan.End()

	// 数据库保存
	job.WithCreateDefault()
	err = db.Repo.Create(ctx, job)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to create job record", zap.Error(err))
		return primitive.NilObjectID, err
	}
	logger.Info("Created job record", zap.String("job_id", job.ID.String()))

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
