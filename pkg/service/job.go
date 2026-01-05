package service

import (
	"context"
	"errors"
	"github.com/bsonger/devflow-common/client/argo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
)

var JobService = &jobService{}

type jobService struct{}

//func NewJobService() *jobService {
//	return &jobService{}
//}

func (s *jobService) Create(ctx context.Context, job *model.Job) (primitive.ObjectID, error) {

	log := logging.LoggerWithContext(ctx)

	// 1️⃣ 获取 Manifest
	manifest, err := ManifestService.Get(ctx, job.ManifestID)
	if err != nil {
		log.Error("Get manifest failed",
			zap.String("manifest_id", job.ManifestID.Hex()),
			zap.Error(err),
		)
		return primitive.NilObjectID, err
	}

	job.ManifestName = manifest.Name
	job.ApplicationId = manifest.ApplicationId

	if job.Type == "" {
		job.Type = model.JobUpgrade
	}

	// 2️⃣ 获取 Application
	app, err := ApplicationService.Get(ctx, manifest.ApplicationId)
	if err != nil {
		log.Error("Get application failed",
			zap.String("application_id", manifest.ApplicationId.Hex()),
			zap.Error(err),
		)
		return primitive.NilObjectID, err
	}
	job.ApplicationName = app.Name
	job.ProjectName = app.ProjectName

	// 3️⃣ 初始化 Job
	job.Status = model.JobPending
	job.WithCreateDefault()

	// 4️⃣ 先落库（非常关键）
	if err := mongo.Repo.Create(ctx, job); err != nil {
		log.Error("Create job record failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log.Info("Job created",
		zap.String("job_id", job.ID.Hex()),
		zap.String("type", job.Type),
	)

	// 5️⃣ 更新 Job → Running
	if err := s.updateStatus(ctx, job.ID, model.JobRunning); err != nil {
		return job.ID, err
	}

	// 6️⃣ 调用 Argo（独立 Span）
	if err := s.syncArgo(ctx, job); err != nil {
		return job.ID, err
	}

	return job.ID, nil
}

func (s *jobService) Get(ctx context.Context, id primitive.ObjectID) (*model.Job, error) {
	app := &model.Job{}
	err := mongo.Repo.FindByID(ctx, app, id)
	return app, err
}

func (s *jobService) Update(ctx context.Context, app *model.Job) error {
	return mongo.Repo.Update(ctx, app)
}

func (s *jobService) Delete(ctx context.Context, id primitive.ObjectID) error {
	app := &model.Job{}
	return mongo.Repo.Delete(ctx, app, id)
}

func (s *jobService) List(ctx context.Context, filter primitive.M) ([]*model.Job, error) {
	var apps []*model.Job
	err := mongo.Repo.List(ctx, &model.Job{}, filter, &apps)
	return apps, err
}

func (s *jobService) updateStatus(ctx context.Context, jobID primitive.ObjectID, status model.JobStatus) error {
	update := primitive.M{
		"$set": primitive.M{
			"status": status,
		},
	}
	return mongo.Repo.UpdateByID(ctx, &model.Job{}, jobID, update)
}

func (s *jobService) syncArgo(ctx context.Context, job *model.Job) error {

	log := logging.LoggerWithContext(ctx)
	var err error
	application := job.GenerateApplication()
	// 3.2 获取当前 trace context
	sc := trace.SpanContextFromContext(ctx)
	application.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}
	application.Labels = map[string]string{
		"status":         string(model.JobRunning),
		model.JobIDLabel: job.ID.Hex(),
	}

	switch job.Type {
	case model.JobInstall:
		err = argo.CreateApplication(ctx, application)
	case model.JobUpgrade, model.JobRollback:
		err = argo.UpdateApplication(ctx, application)
	default:
		err = errors.New("unknown job type")
	}

	if err != nil {
		log.Error("Argo sync failed",
			zap.String("job_id", job.ID.Hex()),
			zap.String("type", job.Type),
			zap.Error(err),
		)
		return err
	}

	log.Info("Argo sync triggered",
		zap.String("job_id", job.ID.Hex()),
	)
	return nil
}
