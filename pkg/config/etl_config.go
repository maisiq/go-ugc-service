package config

import (
	"sync"

	"go.uber.org/zap"
)

var (
	etlConfig *ETLConfig
	etlOnce   sync.Once
)

type ClickhouseConfig struct {
	DSN          string `yaml:"dsn" mapstructure:"dsn"`
	DatabaseName string `yaml:"dbname" mapstructure:"dbname"`
	Username     string `yaml:"username" mapstructure:"username"`
	Password     string `yaml:"password" mapstructure:"password"`
}

type ETLConfig struct {
	App      AppConfig `yaml:"app" mapstructure:"app"`
	Consumer struct {
		GroupID string `yaml:"groupid" mapstructure:"groupid"`
	} `yaml:"consumer" mapstructure:"consumer"`

	Clickhouse ClickhouseConfig `yaml:"clickhouse" mapstructure:"clickhouse"`

	Kafka struct {
		Brokers        []string `yaml:"brokers" mapstructure:"brokers"`
		AnalyticsTopic string   `yaml:"analytics_topic" mapstructure:"analytics_topic"`
	} `yaml:"kafka" mapstructure:"kafka"`
}

func LoadETLConfig(path string) *ETLConfig {
	etlOnce.Do(func() {
		log, _ := zap.NewDevelopment()

		v, err := initViperConfig(path)

		if err != nil {
			log.Fatal("Load config", zap.Error(err))
		}
		v.Unmarshal(&etlConfig)
	})
	return etlConfig
}

func GetETLConfig() *ETLConfig {
	if etlConfig == nil {
		panic("Config not initialized. Call config.LoadETLConfig() first.")
	}
	return etlConfig
}
