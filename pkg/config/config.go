package config

import (
	"fnproxy/pkg/logger"
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `destructure:"server"`
	Target TargetConfig `destructure:"target"`
	Log    LogConfig    `destructure:"log"`
	User   UserConfig   `destructure:"user"`
}

type ServerConfig struct {
	Listen string `destructure:"listen"`
}

type TargetConfig struct {
	Host  string `destructure:"host"`
	Port  int    `destructure:"port"`
	Https bool   `destructure:"https"`
}

type LogConfig struct {
	Level string `destructure:"level"`
}

type UserConfig struct {
	Username string `destructure:"username"`
	Password string `destructure:"password"`
}

var (
	globalCfg *Config
)

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// 设置默认值
	//viper.SetDefault("server.listen", "0.0.0.0:2345")
	//viper.SetDefault("target.host", "10.0.0.115")
	//viper.SetDefault("target.port", 8005)
	//viper.SetDefault("log.level", "info")
	//viper.SetDefault("timeout.read", "30s")
	//viper.SetDefault("timeout.write", "30s")
	//viper.SetDefault("timeout.idle", "120s")

	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf("Error reading config file, %s", err)
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		logger.Errorf("Unable to decode into struct, %v", err)
		return nil, err
	}

	globalCfg = cfg

	return globalCfg, nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalCfg
}
