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

	// 启用环境变量支持
	viper.AutomaticEnv()

	// 绑定环境变量
	viper.BindEnv("server.listen", "SERVER_LISTEN")
	viper.BindEnv("target.host", "TARGET_HOST")
	viper.BindEnv("target.port", "TARGET_PORT")
	viper.BindEnv("target.https", "TARGET_HTTPS")
	viper.BindEnv("user.username", "USER_USERNAME")
	viper.BindEnv("user.password", "USER_PASSWORD")
	viper.BindEnv("log.level", "LOG_LEVEL")

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
