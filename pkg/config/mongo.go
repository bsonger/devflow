package config

type MongoConfig struct {
	URI    string `mapstructure:"uri"`
	DBName string `mapstructure:"db"`
}
