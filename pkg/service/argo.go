package service

import (
	"context"
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/bsonger/devflow/pkg/client"
	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"os"

	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bsonger/devflow/pkg/model"
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
	if err := db.Repo.FindByID(ctx, job, jobID); err != nil {
		logging.LoggerWithContext(ctx).Error("Job not found", zap.String("jobID", jobID.Hex()), zap.Error(err))
		return
	}

	ready := app.Status.Sync.Status == appv1.SyncStatusCodeSynced && app.Status.Health.Status == health.HealthStatusHealthy
	if ready {
		job.Status = model.JobSucceeded
	}
}

// CreateApplication 创建或更新 ArgoCD Application
func CreateApplication(ctx context.Context, job *model.Job) error {
	applications := client.ArgoCdClient.ArgoprojV1alpha1().Applications("argo-cd")
	app := GenerateApplication(ctx, job)

	_, err := applications.Create(ctx, app, metav1.CreateOptions{})
	return err
}

func UpdateApplication(ctx context.Context, job *model.Job) error {
	applications := client.ArgoCdClient.ArgoprojV1alpha1().Applications("argo-cd")
	app := GenerateApplication(ctx, job)
	current, err := applications.Get(ctx, job.ApplicationName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// 3. 保持 name/namespace，替换 spec
	current.Spec = app.Spec
	current.Annotations = app.Annotations
	current.Labels = app.Labels

	// ⚠️ 关键：保留 resourceVersion
	// Kubernetes Update 必须要这个字段
	// current.ResourceVersion 已经是 GET 回来的，直接保留即可。

	// 4. Update
	_, err = applications.Update(ctx, current, metav1.UpdateOptions{})
	return err
}

func GenerateApplication(ctx context.Context, job *model.Job) *appv1.Application {
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
