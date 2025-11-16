package config

type OtelConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
}
