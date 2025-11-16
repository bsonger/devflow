package config

import (
	"github.com/spf13/viper"
)

var C *Config

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Mongo  MongoConfig  `mapstructure:"mongo"`
	Log    LogConfig    `mapstructure:"log"`
	Otel   OtelConfig   `mapstructure:"otel"` // 新增 OTEL
}

func Load() error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("../config/")
	v.AddConfigPath("./config/")

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	if err := v.Unmarshal(&C); err != nil {
		return err
	}

	return nil
}
