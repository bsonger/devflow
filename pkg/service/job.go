package service

import (
	"context"
	"errors"
	"github.com/bsonger/devflow-common/client/argo"
	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var JobService = &jobService{}

type jobService struct{}

//	func NewJobService() *jobService {
//		return &jobService{}
//	}
func (s *jobService) Create(ctx context.Context, job *model.Job) (primitive.ObjectID, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.type", job.Type),
		zap.String("manifest.id", job.ManifestID.Hex()),
	)

	log.Info("job create started")

	// ---------- 1️⃣ 获取 Manifest ----------
	manifest, err := ManifestService.Get(ctx, job.ManifestID)
	if err != nil {
		log.Error("get manifest failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	job.ManifestName = manifest.Name
	job.ApplicationId = manifest.ApplicationId

	// ---------- 2️⃣ 默认值 ----------
	if job.Type == "" {
		job.Type = model.JobUpgrade
	}

	// ---------- 3️⃣ 获取 Application ----------
	app, err := ApplicationService.Get(ctx, manifest.ApplicationId)
	if err != nil {
		log.Error("get application failed",
			zap.String("application.id", manifest.ApplicationId.Hex()),
			zap.Error(err),
		)
		return primitive.NilObjectID, err
	}

	job.ApplicationName = app.Name
	job.ProjectName = app.ProjectName
	job.Env = "prod"

	// ---------- 4️⃣ 初始化 Job ----------
	job.Status = model.JobPending
	job.WithCreateDefault()

	// ---------- 5️⃣ 落库 ----------
	if err := mongo.Repo.Create(ctx, job); err != nil {
		log.Error("create job record failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log = log.With(
		zap.String("job.id", job.ID.Hex()),
		zap.String("application.id", job.ApplicationId.Hex()),
	)

	log.Info("job record created")

	// ---------- 6️⃣ 状态 → Running ----------
	if err := s.updateStatus(ctx, job.ID, model.JobSyncing); err != nil {
		log.Error("update job status to syncing failed", zap.Error(err))
		return job.ID, err
	}

	log.Info("job status changed",
		zap.String("job.status", string(job.Status)),
	)

	// ---------- 7️⃣ 调用 Argo ----------
	if err := s.syncArgo(ctx, job); err != nil {
		s.handleSyncArgoError(ctx, job, err)
		return job.ID, err
	}

	log.Info("job synced to argo successfully")

	return job.ID, nil
}

func (s *jobService) handleSyncArgoError(ctx context.Context, job *model.Job, err error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", job.ID.Hex()),
		zap.String("job.type", job.Type),
	)

	log.Error("sync argo failed", zap.Error(err))

	// 1️⃣ 更新状态 → Failed
	if uErr := s.updateStatus(ctx, job.ID, model.JobSyncFailed); uErr != nil {
		log.Error("update job status to failed failed", zap.Error(uErr))
	}
}

func (s *jobService) Get(ctx context.Context, id primitive.ObjectID) (*model.Job, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", id.Hex()),
		zap.String("operation", "get_job"),
	)

	job := &model.Job{}
	err := mongo.Repo.FindByID(ctx, job, id)
	if err != nil {
		log.Error("get job failed", zap.Error(err))
		return nil, err
	}

	log.Debug("job fetched")
	return job, nil
}

func (s *jobService) Update(ctx context.Context, job *model.Job) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", job.ID.Hex()),
		zap.String("operation", "update_job"),
	)

	if err := mongo.Repo.Update(ctx, job); err != nil {
		log.Error("update job failed", zap.Error(err))
		return err
	}

	log.Debug("job updated")
	return nil
}

func (s *jobService) Delete(ctx context.Context, id primitive.ObjectID) error {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("job.id", id.Hex()),
		zap.String("operation", "delete_job"),
	)

	job := &model.Job{}
	if err := mongo.Repo.Delete(ctx, job, id); err != nil {
		log.Error("delete job failed", zap.Error(err))
		return err
	}

	log.Info("job deleted")
	return nil
}

func (s *jobService) List(ctx context.Context, filter primitive.M) ([]*model.Job, error) {
	log := logging.LoggerWithContext(ctx).With(
		zap.String("operation", "list_jobs"),
		zap.Any("filter", filter),
	)

	var jobs []*model.Job
	if err := mongo.Repo.List(ctx, &model.Job{}, filter, &jobs); err != nil {
		log.Error("list jobs failed", zap.Error(err))
		return nil, err
	}

	log.Debug("list jobs success", zap.Int("count", len(jobs)))
	return jobs, nil
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
