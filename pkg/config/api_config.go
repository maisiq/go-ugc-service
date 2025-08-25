package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	config  *Config
	apiOnce sync.Once
)

type DatabaseConfig struct {
	DSN         string `yaml:"dsn" mapstructure:"dsn"`
	Name        string `yaml:"dbname" mapstructure:"dbname"`
	Collections struct {
		Movies string `yaml:"movies" mapstructure:"movies"`
		Users  string `yaml:"users" mapstructure:"users"`
	} `yaml:"collections" mapstructure:"collections"`
}

type KafkaConfig struct {
	Brokers        []string `yaml:"brokers" mapstructure:"brokers"`
	AnalyticsTopic string   `yaml:"analytics_topic" mapstructure:"analytics_topic"`
}

type CacheConfig struct {
	Addr string `yaml:"addr" mapstructure:"addr"`
}

type SwaggerConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Endpoint string `yaml:"endpoint" mapstructure:"endpoint"`
}

type ServerConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
}

type AppConfig struct {
	Debug        bool `yaml:"debug" mapstructure:"debug"`
	ShutdownTime int  `yaml:"shutdown_time" mapstructure:"shutdown_time"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server" mapstructure:"server"`
	Database DatabaseConfig `yaml:"db" mapstructure:"db"`
	Kafka    KafkaConfig    `yaml:"kafka" mapstructure:"kafka"`
	Cache    CacheConfig    `yaml:"cache" mapstructure:"cache"`
	Swagger  SwaggerConfig  `yaml:"swagger" mapstructure:"swagger"`
	App      AppConfig      `yaml:"app" mapstructure:"app"`
}

func initViperConfig(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(path)

	err := v.ReadInConfig()

	if err != nil {
		return nil, err
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return v, nil
}

func LoadConfig(path string) *Config {
	apiOnce.Do(func() {
		log, _ := zap.NewDevelopment()

		v, err := initViperConfig(path)

		if err != nil {
			log.Fatal("Load config", zap.Error(err))
		}

		v.Unmarshal(&config)

		// Dynamicly get brokers from env (e.g. KAFKA_BROKERS_0) or keep defaults
		var brokers []string

		for i := 0; ; i++ {
			value := v.GetString(fmt.Sprintf("kafka.brokers.%d", i))

			if value == "" {
				break
			}
			brokers = append(brokers, value)
		}

		if len(brokers) > 0 {
			config.Kafka.Brokers = brokers
		}
	})

	return config
}

func GetConfig() *Config {
	if config == nil {
		panic("Config not initialized. Call config.LoadConfig() first.")
	}
	return config
}
