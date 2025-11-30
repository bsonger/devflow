package model

import "k8s.io/client-go/rest"

var C *Config
var KubeConfig *rest.Config

type Config struct {
	Server *ServerConfig `mapstructure:"server"`
	Mongo  *MongoConfig  `mapstructure:"mongo"`
	Log    *LogConfig    `mapstructure:"log"`
	Otel   *OtelConfig   `mapstructure:"otel"` // 新增 OTEL
	Repo   *Repo         `mapstructure:"repo"`
	Consul *Consul       `mapstructure:"consul"`
}

type Consul struct {
	Address string `mapstructure:"address"`
	Key     string `mapstructure:"key"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // console | json
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type MongoConfig struct {
	URI    string `mapstructure:"uri"`
	DBName string `mapstructure:"db"`
}

type OtelConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
}

type Repo struct {
	Address string `json:"address"`
	Path    string `json:"path"`
}
