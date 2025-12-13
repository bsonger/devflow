package service

import (
	"context"
	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/model"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"time"
)

var ManifestService = &manifestService{}

type manifestService struct {
}

func (s *manifestService) CreateManifest(ctx context.Context, m *model.Manifest) (primitive.ObjectID, error) {

	tracer := otel.Tracer("devflow-manifest")
	ctx, span := tracer.Start(ctx, "CreateManifest")
	defer span.End()

	app, err := ApplicationService.Get(ctx, m.ApplicationId)
	if err != nil {
		span.RecordError(err)
		return primitive.NilObjectID, err
	}

	m.GitRepo = app.RepoURL
	m.ApplicationName = app.Name
	m.Name = model.GenerateManifestVersion(app.Name)
	m.Status = model.ManifestPending
	m.WithCreateDefault()

	// 1️⃣ 创建 PipelineRun
	ctx, prSpan := tracer.Start(ctx, "CreatePipelineRun")
	pr, err := CreatePipelineRun(ctx, "devflow-ci", m)
	prSpan.End()
	if err != nil {
		span.RecordError(err)
		return primitive.NilObjectID, err
	}
	m.PipelineID = pr.Name

	// 2️⃣ 查询 Pipeline
	pipeline, err := GetPipeline(
		ctx,
		pr.Namespace,
		pr.Spec.PipelineRef.Name,
	)
	if err != nil {
		span.RecordError(err)
		return primitive.NilObjectID, err
	}

	// 3️⃣ 初始化所有 Step（全部 Pending）
	m.Steps = BuildStepsFromPipeline(pipeline)

	// 4️⃣ 保存 Manifest
	if err := db.Repo.Create(ctx, m); err != nil {
		span.RecordError(err)
		return primitive.NilObjectID, err
	}

	return m.GetID(), nil
}

// GetManifest 根据 ID 查询 Manifest
func (s *manifestService) GetManifest(ctx context.Context, id primitive.ObjectID) (*model.Manifest, error) {
	m := &model.Manifest{}
	err := db.Repo.FindByID(ctx, m, id)
	return m, err
}

// Update UpdateManifest 更新 Manifest
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

	return db.Repo.UpdateOne(ctx, &model.Manifest{}, filter, bson.M{"$set": update})
}

func (s *manifestService) UpdateManifestStatus(ctx context.Context, pipelineID string, status model.ManifestStatus) error {

	filter := bson.M{
		"pipeline_id": pipelineID,
		"status": bson.M{
			"$nin": []model.ManifestStatus{model.ManifestFailed, model.ManifestSucceeded},
		},
	}

	return db.Repo.UpdateOne(
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

	return db.Repo.UpdateOne(
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
	err := db.Repo.FindOne(
		ctx,
		&m,
		bson.M{"pipeline_id": pipelineID},
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
