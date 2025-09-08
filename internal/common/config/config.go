package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `destructure:"server"`
	Target  TargetConfig  `destructure:"target"`
	Log     LogConfig     `destructure:"log"`
	Timeout TimeoutConfig `destructure:"timeout"`
}

type ServerConfig struct {
	Listen string `destructure:"listen"`
}

type TargetConfig struct {
	Host string `destructure:"host"`
	Port int    `destructure:"port"`
}

type LogConfig struct {
	Level string `destructure:"level"`
}

type TimeoutConfig struct {
	Read  time.Duration `destructure:"read"`
	Write time.Duration `destructure:"write"`
	Idle  time.Duration `destructure:"idle"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("server.listen", "0.0.0.0:2345")
	viper.SetDefault("target.host", "10.0.0.115")
	viper.SetDefault("target.port", 8005)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("timeout.read", "30s")
	viper.SetDefault("timeout.write", "30s")
	viper.SetDefault("timeout.idle", "120s")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
