package config

import (
	"context"
	"fmt"
	"github.com/bsonger/devflow/pkg/client"
	"github.com/bsonger/devflow/pkg/db"
	"github.com/bsonger/devflow/pkg/logging"
	"github.com/bsonger/devflow/pkg/model"
	"github.com/bsonger/devflow/pkg/otel"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func Load() error {
	v := viper.New()
	//v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config/")
	v.AddConfigPath("/etc/devflow/config/")

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	if err := v.Unmarshal(&model.C); err != nil {
		return err
	}
	var err error
	model.KubeConfig, err = LoadKubeConfig()
	if err != nil {
		return err
	}
	reloadConfig()
	err = InitConsulClient(model.C.Consul)
	if err != nil {
		return err
	}
	LoadConsulConfigAndMerge(model.C.Consul)
	WatchConsul(model.C.Consul)
	return nil
}

func reloadConfig() {
	logging.Init()

	ctx := context.Background()
	otel.Init(model.C.Otel.Endpoint, model.C.Otel.ServiceName)

	_, err := db.InitMongo(ctx, model.C.Mongo.URI, model.C.Mongo.DBName, logging.Logger)
	if err != nil {
		logging.Logger.Fatal("mongo init failed", zap.Error(err))
	}

	err = client.InitTektonClient()
	if err != nil {
		logging.Logger.Fatal("tekton init failed", zap.Error(err))
	}
	err = client.InitArgoCdClient()

	if err != nil {
		logging.Logger.Fatal("argo init failed", zap.Error(err))
	}
}

// LoadKubeConfig 自动加载 kubeconfig（本地）或 InCluster（Pod 内）
func LoadKubeConfig() (*rest.Config, error) {
	// 1. 尝试本地 kubeconfig
	if cfg, err := loadLocalKubeConfig(); err == nil {
		return cfg, nil
	}

	// 2. 回退到 in-cluster 配置
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	return nil, fmt.Errorf("failed to load kubeconfig and in-cluster config")
}

// loadLocalKubeConfig 从 $HOME/.kube/config 加载
func loadLocalKubeConfig() (*rest.Config, error) {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows fallback
	}

	kubeconfig := filepath.Join(home, ".kube", "config")

	// 如果文件不存在，直接返回 error
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, err
	}

	// 使用 kubeconfig 构建 config
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
