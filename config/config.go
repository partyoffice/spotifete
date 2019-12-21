package config

import (
	"github.com/spf13/viper"
)

var config *viper.Viper

func GetConfig() *viper.Viper {
	if config == nil {
		var err error
		config = viper.New()
		config.SetConfigType("yaml")
		config.SetConfigName("spotifete-config")
		config.AddConfigPath("/etc/spotifete")
		config.AddConfigPath(".")
		err = config.ReadInConfig()
		if err != nil {
			panic("Could not read config file.")
		}
	}

	return config
}
