package service

import (
	"context"
	"github.com/bsonger/devflow-common/client/tekton"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"time"

	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
)

var ManifestService = &manifestService{}

const (
	namespace = "tekton-pipelines"
)

type manifestService struct {
}

func (s *manifestService) CreateManifest(ctx context.Context, m *model.Manifest) (primitive.ObjectID, error) {
	logger := logging.LoggerFromContext(ctx)

	// ---- Entry log（一次请求的主线）----
	logger.Info("create manifest start",
		zap.String("application_id", m.ApplicationId.Hex()),
		zap.String("branch", m.Branch),
	)

	// 1️⃣ 获取 Application
	app, err := ApplicationService.Get(ctx, m.ApplicationId)
	if err != nil {
		logger.Error("get application failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	logger.Debug("application loaded",
		zap.String("application", app.Name),
		zap.String("repo", app.RepoURL),
	)

	// 2️⃣ 初始化 Manifest 基础信息
	m.GitRepo = app.RepoURL
	m.ApplicationName = app.Name
	m.Replica = app.Replica
	m.Service = app.Service
	m.Type = app.Type
	m.Envs = app.Envs
	m.ConfigMaps = app.ConfigMaps
	m.Internet = app.Internet

	m.Name = model.GenerateManifestVersion(app.Name)
	m.Status = model.ManifestPending
	m.WithCreateDefault()
	if m.Branch == "" {
		m.Branch = "main"
	}

	logger.Debug("manifest initialized",
		zap.String("manifest", m.Name),
		zap.String("application", app.Name),
	)

	// ---- PVC ----
	pvc, err := tekton.CreatePVC(ctx, namespace, "devflow-ci", "local-path", "1Gi")
	if err != nil {
		logger.Error("create pvc failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	logger.Debug("pvc created", zap.String("pvc", pvc.Name))

	// 3️⃣ PipelineRun Span
	pctx, span := devflowTracer.Start(ctx, "Tekton.CreatePipelineRun")
	defer span.End()

	// 3.1 PipelineRun
	pr := m.GeneratePipelineRun("devflow-ci", pvc.Name)

	sc := trace.SpanContextFromContext(pctx)
	pr.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}

	// 3.2 创建 PipelineRun
	pr, err = tekton.CreatePipelineRun(pctx, namespace, pr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error("create pipelineRun failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	logger.Info("pipelineRun created",
		zap.String("pipelineRun", pr.Name),
		zap.String("pipeline", pr.Spec.PipelineRef.Name),
	)

	// ---- PVC Owner ----
	if err := tekton.PatchPVCOwner(ctx, pvc, pr); err != nil {
		logger.Warn("patch pvc owner failed", zap.Error(err))
	}

	m.PipelineID = pr.Name

	// 4️⃣ 查询 Pipeline
	pipeline, err := tekton.GetPipeline(ctx, pr.Namespace, pr.Spec.PipelineRef.Name)
	if err != nil {
		logger.Error("get pipeline failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	logger.Debug("pipeline fetched", zap.String("pipeline", pipeline.Name))

	// 5️⃣ 初始化 Step
	m.Steps = BuildStepsFromPipeline(pipeline)

	logger.Debug("steps initialized", zap.Int("step_count", len(m.Steps)))

	// 6️⃣ 保存 Manifest
	if err := mongo.Repo.Create(ctx, m); err != nil {
		logger.Error("save manifest failed", zap.Error(err))
		return primitive.NilObjectID, err
	}

	logger.Info("create manifest success",
		zap.String("manifest", m.Name),
		zap.String("pipelineRun", m.PipelineID),
	)

	return m.GetID(), nil
}

// GetManifest 根据 ID 查询 Manifest
func (s *manifestService) GetManifest(ctx context.Context, id primitive.ObjectID) (*model.Manifest, error) {
	logger := logging.LoggerWithContext(ctx)

	logger.Debug("get manifest start",
		zap.String("manifest_id", id.Hex()),
	)

	m := &model.Manifest{}
	if err := mongo.Repo.FindByID(ctx, m, id); err != nil {
		logger.Error("get manifest failed",
			zap.String("manifest_id", id.Hex()),
			zap.Error(err),
		)
		return nil, err
	}

	logger.Debug("get manifest success",
		zap.String("manifest_id", id.Hex()),
		zap.String("manifest_name", m.Name),
		zap.String("status", string(m.Status)),
	)

	return m, nil
}

// Update UpdateManifest 更新 Manifest
func (s *manifestService) Update(ctx context.Context, m *model.Manifest) error {

	logger := logging.LoggerWithContext(ctx)

	logger.Debug("update manifest start",
		zap.String("manifest_id", m.GetID().Hex()),
		zap.String("manifest_name", m.Name),
		zap.String("status", string(m.Status)),
	)

	if err := mongo.Repo.Update(ctx, m); err != nil {
		logger.Error("update manifest failed",
			zap.String("manifest_id", m.GetID().Hex()),
			zap.String("manifest_name", m.Name),
			zap.Error(err),
		)
		return err
	}

	logger.Info("update manifest success",
		zap.String("manifest_id", m.GetID().Hex()),
		zap.String("manifest_name", m.Name),
		zap.String("status", string(m.Status)),
	)

	return nil
}
func (s *manifestService) List(ctx context.Context, filter primitive.M) ([]model.Manifest, error) {

	logger := logging.LoggerWithContext(ctx)

	logger.Debug("list manifests start")

	var manifests []model.Manifest
	if err := mongo.Repo.List(ctx, &model.Manifest{}, filter, &manifests); err != nil {
		logger.Error("list manifests failed",
			zap.Error(err),
		)
		return nil, err
	}

	logger.Debug("list manifests success",
		zap.Int("count", len(manifests)),
	)

	return manifests, nil
}

func (s *manifestService) Get(ctx context.Context, id primitive.ObjectID) (*model.Manifest, error) {
	app := &model.Manifest{}
	err := mongo.Repo.FindByID(ctx, app, id)
	return app, err
}

func (s *manifestService) UpdateStepStatus(ctx context.Context, pipelineID, taskName string, status model.StepStatus, message string, start, end *time.Time) error {

	update := bson.M{
		"steps.$.status":  status,
		"steps.$.message": message,
		"updated_at":      time.Now(),
	}

	if start != nil {
		update["steps.$.start_time"] = start
	}
	if end != nil {
		update["steps.$.end_time"] = end
	}

	filter := bson.M{
		"pipeline_id": pipelineID,
		"steps": bson.M{
			"$elemMatch": bson.M{
				"task_name": taskName,
				"status": bson.M{
					"$nin": []model.StepStatus{model.StepFailed, model.StepSucceeded},
				},
			},
		},
	}

	return mongo.Repo.UpdateOne(ctx, &model.Manifest{}, filter, bson.M{"$set": update})
}

func (s *manifestService) UpdateManifestStatus(ctx context.Context, pipelineID string, status model.ManifestStatus) error {

	filter := bson.M{
		"pipeline_id": pipelineID,
		"status": bson.M{
			"$nin": []model.ManifestStatus{model.ManifestFailed, model.ManifestSucceeded},
		},
	}

	return mongo.Repo.UpdateOne(
		ctx,
		&model.Manifest{},
		filter,
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)
}

func BuildStepsFromPipeline(pipeline *v1.Pipeline) []model.ManifestStep {

	steps := make([]model.ManifestStep, 0)

	for _, task := range pipeline.Spec.Tasks {
		steps = append(steps, model.ManifestStep{
			TaskName: task.Name,
			Status:   model.StepPending,
		})
	}

	for _, task := range pipeline.Spec.Finally {
		steps = append(steps, model.ManifestStep{
			TaskName: task.Name,
			Status:   model.StepPending,
		})
	}

	return steps
}

func (s *manifestService) BindTaskRun(ctx context.Context, pipelineID, taskName, taskRun string) error {

	return mongo.Repo.UpdateOne(
		ctx,
		&model.Manifest{},
		bson.M{
			"pipeline_id":     pipelineID,
			"steps.task_name": taskName,
		},
		bson.M{
			"$set": bson.M{
				"steps.$.task_run": taskRun,
				"updated_at":       time.Now(),
			},
		},
	)
}

func (s *manifestService) GetManifestByPipelineID(ctx context.Context, pipelineID string) (*model.Manifest, error) {

	var m model.Manifest
	err := mongo.Repo.FindOne(
		ctx,
		&m,
		bson.M{"pipeline_id": pipelineID},
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
