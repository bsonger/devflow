package service

import (
	"context"
	"errors"
	"fmt"
	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/bsonger/devflow-common/client/argo"
	"go.opentelemetry.io/otel/trace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
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
	tracer := otel.Tracer("devflow/job")
	ctx, span := tracer.Start(ctx, "JobService.Create")
	defer span.End()

	log := logging.LoggerWithContext(ctx)

	// 1️⃣ 获取 Manifest
	manifest, err := ManifestService.Get(ctx, job.ManifestID)
	if err != nil {
		span.RecordError(err)
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
		span.RecordError(err)
		log.Error("Get application failed",
			zap.String("application_id", manifest.ApplicationId.Hex()),
			zap.Error(err),
		)
		return primitive.NilObjectID, err
	}
	job.ApplicationName = app.Name

	// 3️⃣ 初始化 Job
	job.Status = model.JobPending
	job.WithCreateDefault()

	// 4️⃣ 先落库（非常关键）
	if err := mongo.Repo.Create(ctx, job); err != nil {
		span.RecordError(err)
		log.Error("Create job record failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	log.Info("Job created",
		zap.String("job_id", job.ID.Hex()),
		zap.String("type", job.Type),
	)

	// 5️⃣ 更新 Job → Running
	if err := s.updateStatus(ctx, job.ID, model.JobRunning); err != nil {
		span.RecordError(err)
		return job.ID, err
	}

	// 6️⃣ 调用 Argo（独立 Span）
	if err := s.syncArgo(ctx, job); err != nil {
		span.RecordError(err)
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
	tracer := otel.Tracer("devflow/job")
	ctx, span := tracer.Start(ctx, "Argo.Sync")
	defer span.End()

	log := logging.LoggerWithContext(ctx)
	var err error
	application := s.GenerateApplication(ctx, job)
	switch job.Type {
	case model.JobInstall:
		err = argo.CreateApplication(ctx, application)
	case model.JobUpgrade, model.JobRollback:
		err = argo.UpdateApplication(ctx, application)
	default:
		err = errors.New("unknown job type")
	}

	if err != nil {
		span.RecordError(err)
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

func (s *jobService) GenerateApplication(ctx context.Context, job *model.Job) *appv1.Application {
	env := os.Getenv("env")
	var path string

	if env != "" {
		path = fmt.Sprintf("%s/%s/overlays/%s", job.ApplicationName, job.ManifestName, os.Getenv("env"))
	} else {
		path = fmt.Sprintf("%s/%s/base", job.ApplicationName, job.ManifestName)
	}

	span := trace.SpanFromContext(ctx)
	labels := map[string]string{
		"devflow/job-id": job.ID.Hex(),
	}
	if span.SpanContext().IsValid() {
		labels["trace_id"] = span.SpanContext().TraceID().String()
	}

	app := &appv1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.ApplicationName,
			Namespace: "argo-cd",
			Labels:    labels,
		},
		Spec: appv1.ApplicationSpec{
			Project: "default",
			Source: &appv1.ApplicationSource{
				RepoURL:        model.C.Repo.Address,
				TargetRevision: "main",
				Path:           path,
				//Kustomize: &appv1.ApplicationSourceKustomize{
				//	// 可以设置 namePrefix, images, 带 patch 的 kustomize 等
				//	//CommonLabels: job.CommonLabels,
				//},
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "apps",
			},
			SyncPolicy: &appv1.SyncPolicy{
				Automated: &appv1.SyncPolicyAutomated{
					Prune:    true, // 自动删除
					SelfHeal: true, // 自动修复漂移
				},
			},
		},
	}
	return app
}
