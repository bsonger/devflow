package service

import (
	"context"
	"github.com/argoproj/gitops-engine/pkg/health"

	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

func StartArgoCdInformer(ctx context.Context) error {
	return nil
}

func handleArgoEvent(ctx context.Context, obj interface{}) {
	app, ok := obj.(*appv1.Application)
	if !ok {
		logging.LoggerWithContext(ctx).Error("invalid object type")
		return
	}

	jobIDStr, ok := app.Labels["devflow/job-id"]
	if !ok || jobIDStr == "" {
		logging.LoggerWithContext(ctx).Warn("jobID label missing")
		return
	}

	jobID, err := primitive.ObjectIDFromHex(jobIDStr)
	if err != nil {
		logging.LoggerWithContext(ctx).Error("invalid jobID format", zap.String("jobID", jobIDStr), zap.Error(err))
		return
	}

	job := &model.Job{}
	if err := mongo.Repo.FindByID(ctx, job, jobID); err != nil {
		logging.LoggerWithContext(ctx).Error("Job not found", zap.String("jobID", jobID.Hex()), zap.Error(err))
		return
	}

	ready := app.Status.Sync.Status == appv1.SyncStatusCodeSynced && app.Status.Health.Status == health.HealthStatusHealthy
	if ready {
		job.Status = model.JobSucceeded
	}
}

//// CreateApplication 创建或更新 ArgoCD Application
//func CreateApplication(ctx context.Context, job *model.Job) error {
//	applications := argo.ArgoCdClient.ArgoprojV1alpha1().Applications("argo-cd")
//	app := GenerateApplication(ctx, job)
//
//	_, err := applications.Create(ctx, app, metav1.CreateOptions{})
//	return err
//}

//func UpdateApplication(ctx context.Context, job *model.Job) error {
//	applications := client.ArgoCdClient.ArgoprojV1alpha1().Applications("argo-cd")
//	app := GenerateApplication(ctx, job)
//	current, err := applications.Get(ctx, job.ApplicationName, metav1.GetOptions{})
//	if err != nil {
//		return err
//	}
//
//	// 3. 保持 name/namespace，替换 spec
//	current.Spec = app.Spec
//	current.Annotations = app.Annotations
//	current.Labels = app.Labels
//
//	// ⚠️ 关键：保留 resourceVersion
//	// Kubernetes Update 必须要这个字段
//	// current.ResourceVersion 已经是 GET 回来的，直接保留即可。
//
//	// 4. Update
//	_, err = applications.Update(ctx, current, metav1.UpdateOptions{})
//	return err
//}
