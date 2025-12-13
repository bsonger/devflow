package model

import "k8s.io/client-go/rest"

var C *Config
var KubeConfig *rest.Config

type Config struct {
	Server *ServerConfig `mapstructure:"server" json:"server" yaml:"server"`
	Mongo  *MongoConfig  `mapstructure:"mongo"  json:"mongo"  yaml:"mongo"`
	Log    *LogConfig    `mapstructure:"log"    json:"log"    yaml:"log"`
	Otel   *OtelConfig   `mapstructure:"otel"   json:"otel"   yaml:"otel"`
	Repo   *Repo         `mapstructure:"repo"   json:"repo"   yaml:"repo"`
	Consul *Consul       `mapstructure:"consul" json:"consul" yaml:"consul"`
}

type Consul struct {
	Address string `mapstructure:"address" json:"address" yaml:"address"`
	Key     string `mapstructure:"key"     json:"key"     yaml:"key"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"  json:"level"  yaml:"level"`
	Format string `mapstructure:"format" json:"format" yaml:"format"` // console | json
}

type ServerConfig struct {
	Port int `mapstructure:"port" json:"port" yaml:"port"`
}

type MongoConfig struct {
	URI    string `mapstructure:"uri" json:"uri" yaml:"uri"`
	DBName string `mapstructure:"db"  json:"db"  yaml:"db"`
}

type OtelConfig struct {
	Endpoint    string `mapstructure:"endpoint"     json:"endpoint"     yaml:"endpoint"`
	ServiceName string `mapstructure:"service_name" json:"service_name" yaml:"service_name"`
}

type Repo struct {
	Address string `mapstructure:"address" json:"address" yaml:"address"`
	Path    string `mapstructure:"path"    json:"path"    yaml:"path"`
}
