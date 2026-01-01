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

	// 1️⃣ 获取 Application
	app, err := ApplicationService.Get(ctx, m.ApplicationId)
	if err != nil {
		return primitive.NilObjectID, err
	}
	logging.LoggerWithContext(ctx).Debug("get application", zap.String("name", app.Name))

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

	pvc, err := tekton.CreatePVC(ctx, namespace, "devflow-ci", "local-path", "1Gi")
	if err != nil {
		return primitive.NilObjectID, err
	}

	// 3️⃣ 创建 PipelineRun Span
	pctx, span := devflowTracer.Start(ctx, "Tekton.CreatePipelineRun")

	// 3.1 生成 PipelineRun 对象
	pr := m.GeneratePipelineRun("devflow-ci", pvc.Name)

	// 3.2 获取当前 trace context
	sc := trace.SpanContextFromContext(pctx)
	pr.Annotations = map[string]string{
		model.TraceIDAnnotation: sc.TraceID().String(),
		model.SpanAnnotation:    sc.SpanID().String(),
	}

	// 3.3 调用 Tekton 创建 PipelineRun
	pr, err = tekton.CreatePipelineRun(ctx, namespace, pr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return primitive.NilObjectID, err
	}

	logging.LoggerWithContext(ctx).Debug("pipeline run created",
		zap.String("pipelineRun", pr.Name),
		zap.String("pipeline", pr.Spec.PipelineRef.Name),
	)
	span.End()

	err = tekton.PatchPVCOwner(ctx, pvc, pr)
	if err != nil {
		logging.LoggerWithContext(ctx).Error("patch pvc owner failed", zap.Error(err))
	}

	m.PipelineID = pr.Name

	// 4️⃣ 查询 Pipeline 定义
	pipeline, err := tekton.GetPipeline(ctx, pr.Namespace, pr.Spec.PipelineRef.Name)
	if err != nil {
		return primitive.NilObjectID, err
	}
	logging.LoggerWithContext(ctx).Debug("pipeline fetched", zap.String("pipeline", pipeline.Name))

	// 5️⃣ 初始化所有 Step（全部 Pending）
	m.Steps = BuildStepsFromPipeline(pipeline)

	// 6️⃣ 保存 Manifest 到数据库
	if err := mongo.Repo.Create(ctx, m); err != nil {
		return primitive.NilObjectID, err
	}
	logging.LoggerWithContext(ctx).Debug("manifest saved", zap.String("manifest", m.Name))

	return m.GetID(), nil
}

// GetManifest 根据 ID 查询 Manifest
func (s *manifestService) GetManifest(ctx context.Context, id primitive.ObjectID) (*model.Manifest, error) {
	m := &model.Manifest{}
	err := mongo.Repo.FindByID(ctx, m, id)
	return m, err
}

// Update UpdateManifest 更新 Manifest
func (s *manifestService) Update(ctx context.Context, m *model.Manifest) error {
	return mongo.Repo.Update(ctx, m)
}

func (s *manifestService) List(ctx context.Context, filter primitive.M) ([]model.Manifest, error) {
	var apps []model.Manifest
	err := mongo.Repo.List(ctx, &model.Manifest{}, filter, &apps)
	return apps, err
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
