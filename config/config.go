package config

import (
	"github.com/google/logger"
	"github.com/spf13/viper"
)

var config *viper.Viper

func GetConfig() *viper.Viper {
	if config == nil {
		config = viper.New()
		config.SetConfigType("yaml")
		config.SetConfigName("spotifete-config")
		config.AddConfigPath("/etc/spotifete")
		config.AddConfigPath(".")
		err := config.ReadInConfig()
		if err != nil {
			logger.Fatal("Could not read config file.")
		}
	}

	return config
}
