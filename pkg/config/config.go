package config

import (
	"fmt"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

var C *Config
var KubeConfig *rest.Config

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Mongo  MongoConfig  `mapstructure:"mongo"`
	Log    LogConfig    `mapstructure:"log"`
	Otel   OtelConfig   `mapstructure:"otel"` // 新增 OTEL
	Repo   Repo         `mapstructure:"repo"`
}

func Load() error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("../config/")
	v.AddConfigPath("./config/")
	v.AddConfigPath("/etc/devflow/config/")

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	if err := v.Unmarshal(&C); err != nil {
		return err
	}
	var err error
	KubeConfig, err = LoadKubeConfig()
	if err != nil {
		return err
	}
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
