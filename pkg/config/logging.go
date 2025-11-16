package config

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // console | json
}
