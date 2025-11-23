package argo

import (
	"context"
	"fmt"
	"os"

	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	argoclient "github.com/argoproj/argo-cd/v3/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bsonger/devflow/pkg/config"
	"github.com/bsonger/devflow/pkg/model"
)

var argoCdClient *argoclient.Clientset

// InitArgocdClient 初始化 ArgoCD client
func InitArgocdClient() error {
	var err error
	argoCdClient, err = argoclient.NewForConfig(config.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create argo cd client: %w", err)
	}
	return nil
}

// CreateApplication 创建或更新 ArgoCD Application
func CreateApplication(ctx context.Context, job *model.Job) error {
	applications := argoCdClient.ArgoprojV1alpha1().Applications("argo-cd")
	app := GenerateApplication(ctx, job)

	_, err := applications.Create(ctx, app, metav1.CreateOptions{})
	return err
}

func UpdateApplication(ctx context.Context, job *model.Job) error {
	applications := argoCdClient.ArgoprojV1alpha1().Applications("argo-cd")
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

	app := &appv1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.ApplicationName,
			Namespace: "argo-cd",
		},
		Spec: appv1.ApplicationSpec{
			Project: "default",
			Source: &appv1.ApplicationSource{
				RepoURL:        config.C.Repo.Address,
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
