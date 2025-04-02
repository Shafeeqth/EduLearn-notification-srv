package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	KafkaBrokers  []string
	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	DatabaseDSN   string
	RedisAddr     string
	ConsumerGroup string
	GRpcPort      string
}

func LoadConfig(logger *zap.Logger) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		return nil, err
	}

	cfg := &Config{
		KafkaBrokers:  viper.GetStringSlice("kafka.brokers"),
		SMTPHost:      viper.GetString("smtp.host"),
		SMTPPort:      viper.GetString("smtp.port"),
		SMTPUsername:  viper.GetString("smtp.username"),
		SMTPPassword:  viper.GetString("smtp.password"),
		DatabaseDSN:   viper.GetString("database.dsn"),
		RedisAddr:     viper.GetString("redis.addr"),
		GRpcPort:      viper.GetString("grpc.port"),
		ConsumerGroup: viper.GetString("kafka.consumer_group"),
	}

	return cfg, nil
}
