package service

import (
	"context"
	"fmt"
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/otel"
	"go.uber.org/zap"
	"time"

	tknv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	informers "github.com/tektoncd/pipeline/pkg/client/informers/externalversions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/apis"

	"github.com/bsonger/devflow/pkg/client"
	"github.com/bsonger/devflow/pkg/model"
)

func Start(ctx context.Context) error {
	factory := informers.NewSharedInformerFactory(client.TektonClient, 0)

	prInformer := factory.Tekton().V1().PipelineRuns().Informer()
	trInformer := factory.Tekton().V1().TaskRuns().Informer()

	prInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: onPipelineRun,
		UpdateFunc: func(oldObj, newObj interface{}) {
			if oldObj == newObj {
				return
			}
			onPipelineRun(newObj)
		},
	})

	trInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: onTaskRun,
		UpdateFunc: func(oldObj, newObj interface{}) {
			if oldObj == newObj {
				return
			}
			onTaskRun(newObj)
		},
	})

	go factory.Start(ctx.Done())

	// 等待缓存同步
	cache.WaitForCacheSync(ctx.Done(),
		prInformer.HasSynced,
		trInformer.HasSynced,
	)

	return nil
}

func onTaskRun(obj interface{}) {
	tr := obj.(*v1.TaskRun)
	ctx := context.Background()
	logging.Logger.Debug("TaskRun event",
		zap.String("taskRun", tr.Name),
		zap.String("pipelineRun", tr.Labels["tekton.dev/pipelineRun"]),
		zap.String("pipelineTask", tr.Labels["tekton.dev/pipelineTask"]),
		zap.String("taskRef", tr.Spec.TaskRef.Name),
	)

	pipelineID := tr.Labels["tekton.dev/pipelineRun"]
	taskRun := tr.Name
	taskName := tr.Labels["tekton.dev/pipelineTask"]
	manifest, err := ManifestService.GetManifestByPipelineID(ctx, pipelineID)

	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Failed to get Manifest by PipelineID failed: %s", pipelineID), zap.Error(err))
		return
	}
	manifestStep := manifest.GetStep(taskName)

	if manifestStep == nil || manifestStep.Status == model.StepFailed || manifestStep.Status == model.StepSucceeded {
		logging.Logger.Info(fmt.Sprintf("Skipping manifest %s step %s", pipelineID, taskName))
		return
	}

	// 1️⃣ 绑定 TaskRun
	_ = ManifestService.BindTaskRun(
		ctx, pipelineID, taskName, taskRun,
	)

	cond := tr.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		return
	}

	switch cond.Status {
	case corev1.ConditionUnknown:
		start := tr.Status.StartTime.Time
		_ = ManifestService.UpdateStepStatus(ctx, pipelineID, taskName, model.StepRunning, cond.Message, &start, nil)

	case corev1.ConditionTrue:
		end := tr.Status.CompletionTime.Time
		_ = ManifestService.UpdateStepStatus(ctx, pipelineID, taskName, model.StepSucceeded, cond.Message, nil, &end)

	case corev1.ConditionFalse:
		end := tr.Status.CompletionTime.Time
		_ = ManifestService.UpdateStepStatus(ctx, pipelineID, taskName, model.StepFailed, cond.Message, nil, &end)
	}
}

func onPipelineRun(obj interface{}) {
	pr := obj.(*v1.PipelineRun)
	ctx := context.Background()

	pipelineID := pr.Name

	manifest, err := ManifestService.GetManifestByPipelineID(ctx, pipelineID)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Failed to get Manifest by PipelineID failed: %s", pipelineID), zap.Error(err))
		return
	}
	if manifest.Status == model.ManifestFailed || manifest.Status == model.ManifestSucceeded {
		logging.Logger.Info(fmt.Sprintf("Skipping manifest step %s", pipelineID))
		return
	}
	cond := pr.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		return
	}

	switch cond.Status {
	case corev1.ConditionUnknown:
		_ = ManifestService.UpdateManifestStatus(ctx, pipelineID, model.ManifestRunning)
	case corev1.ConditionTrue:
		_ = ManifestService.UpdateManifestStatus(ctx, pipelineID, model.ManifestSucceeded)
	case corev1.ConditionFalse:
		_ = ManifestService.UpdateManifestStatus(ctx, pipelineID, model.ManifestFailed)
	}
}

func GetPipeline(ctx context.Context, namespace string, name string) (*v1.Pipeline, error) {
	return client.TektonClient.TektonV1().Pipelines(namespace).Get(ctx, name, metav1.GetOptions{})
}

func CreatePipelineRun(ctx context.Context, pipelineName string, manifest *model.Manifest) (*tknv1.PipelineRun, error) {
	ctx, span := otel.Start(ctx, "createPipelineRun")
	defer span.End()
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

	created, err := client.TektonClient.TektonV1().PipelineRuns("tekton-pipelines").Create(context.TODO(), pipelineRun, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return created, err
}
