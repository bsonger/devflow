package tekton

import (
	"context"
	"fmt"
	"github.com/bsonger/devflow/pkg/config"
	"github.com/bsonger/devflow/pkg/model"
	"time"

	tknv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var TektonClient *tektonclient.Clientset

func InitTektonClient() error {
	var err error
	TektonClient, err = tektonclient.NewForConfig(config.KubeConfig)
	if err != nil {
		return err
	}
	return nil
}

func CreatePipelineRun(ctx context.Context, pipelineName string, manifest *model.Manifest) (*tknv1.PipelineRun, error) {

	// 随机生成一个 PipelineRun 名称
	prName := fmt.Sprintf("%s-run-%d", pipelineName, time.Now().Unix())

	// 构造 PipelineRun 参数
	prParams := []tknv1.Param{
		{
			Name: "git-url",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: manifest.GitRepo,
			},
		},
		{
			Name: "git-revision",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: manifest.Branch,
			},
		},
		{
			Name: "image-registry",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: "registry.cn-hangzhou.aliyuncs.com/devflow",
			},
		},
		{
			Name: "name",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: manifest.ApplicationName,
			},
		},
		{
			Name: "image-tag",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: manifest.Version,
			},
		},
		{
			Name: "manifest-name",
			Value: tknv1.ParamValue{
				Type:      tknv1.ParamTypeString,
				StringVal: manifest.Name,
			},
		},
	}

	// 构造 PipelineRun 对象
	pipelineRun := &tknv1.PipelineRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineRun",
			APIVersion: "tekton.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      prName,
			Namespace: "tekton-pipelines",
		},
		Spec: tknv1.PipelineRunSpec{
			PipelineRef: &tknv1.PipelineRef{
				Name: pipelineName,
			},
			Params: prParams,
			Workspaces: []tknv1.WorkspaceBinding{
				{
					Name: "source",
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "git-source-pvc",
					},
				},
				{
					Name: "dockerconfig",
					Secret: &corev1.SecretVolumeSource{
						SecretName: "aliyun-docker-config",
					},
				},
				{
					Name: "ssh",
					Secret: &corev1.SecretVolumeSource{
						SecretName: "git-ssh-secret",
					},
				},
			},
		},
	}

	// 创建 PipelineRun

	created, err := TektonClient.TektonV1().PipelineRuns("tekton-pipelines").Create(context.TODO(), pipelineRun, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return created, err
}
