package argo

import (
	"context"
	"fmt"
	appv1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	argoclient "github.com/argoproj/argo-cd/v3/pkg/client/clientset/versioned"
	"github.com/bsonger/devflow/pkg/config"
	"github.com/bsonger/devflow/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				Path:           fmt.Sprintf("%s/%s", config.C.Repo.Path, job.ApplicationName),
				//Kustomize: &appv1.ApplicationSourceKustomize{
				//	// 可以设置 namePrefix, images, 带 patch 的 kustomize 等
				//	//CommonLabels: job.CommonLabels,
				//},
			},
			Destination: appv1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: job.ProjectName,
			},
			SyncPolicy: &appv1.SyncPolicy{
				Automated: &appv1.SyncPolicyAutomated{
					Prune:    true, // 自动删除
					SelfHeal: true, // 自动修复漂移
				},
			},
		},
	}

	_, err := applications.Create(ctx, app, metav1.CreateOptions{})
	return err
}

func UpdateApplication(ctx context.Context, job *model.Job) error {
	applications := argoCdClient.ArgoprojV1alpha1().Applications("argo-cd")
	app := &appv1.Application{}
	_, err := applications.Update(ctx, app, metav1.UpdateOptions{})
	return err
}
