package config

import (
	"context"
	"fmt"
	"github.com/bsonger/devflow-common/client/argo"
	"github.com/bsonger/devflow-common/client/logging"
	"github.com/bsonger/devflow-common/client/mongo"
	"github.com/bsonger/devflow-common/client/otel"
	"github.com/bsonger/devflow-common/client/tekton"
	"github.com/bsonger/devflow-common/model"

	"github.com/spf13/viper"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"

	"github.com/bsonger/devflow-common/client/consul"
)

type Config struct {
	Server *model.ServerConfig `mapstructure:"server" json:"server" yaml:"server"`
	Mongo  *model.MongoConfig  `mapstructure:"mongo"  json:"mongo"  yaml:"mongo"`
	Log    *model.LogConfig    `mapstructure:"log"    json:"log"    yaml:"log"`
	Otel   *model.OtelConfig   `mapstructure:"otel"   json:"otel"   yaml:"otel"`
	Repo   *model.Repo         `mapstructure:"repo"   json:"repo"   yaml:"repo"`
	Consul *model.Consul       `mapstructure:"consul" json:"consul" yaml:"consul"`
}

func Load() (*Config, error) {
	v := viper.New()
	//v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config/")
	v.AddConfigPath("/etc/devflow/config/")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config *Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}
	var err error
	model.KubeConfig, err = LoadKubeConfig()
	if err != nil {
		return nil, err
	}
	err = consul.InitConsulClient(config.Consul)
	if err != nil {
		return nil, err
	}
	//consul.LoadConsulConfigAndMerge(config.Consul)

	reloadConfig(context.Background(), config)

	return config, nil
}

func reloadConfig(ctx context.Context, config *Config) error {
	fmt.Println(config.Log)
	logging.InitZapLogger(ctx, config.Log)
	_, err := otel.InitOtel(ctx, config.Otel)
	if err != nil {
		return err
	}

	mongo.InitMongo(ctx, config.Mongo, logging.Logger)
	kubeconfig, err := LoadKubeConfig()
	tekton.InitTektonClient(ctx, kubeconfig, logging.Logger)
	argo.InitArgoCdClient(kubeconfig)
	model.InitConfigRepo(config.Repo)
	return nil
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
