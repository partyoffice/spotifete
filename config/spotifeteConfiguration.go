package config

import (
	"github.com/spf13/viper"
)

type spotifeteConfiguration struct {
	BaseUrl          string
	Port             int
	ReleaseMode      bool
	LogDirectory     string
	AppConfiguration appConfiguration
}

func (c spotifeteConfiguration) read(viperConfiguration *viper.Viper) spotifeteConfiguration {
	c.BaseUrl = getRequiredString(viperConfiguration, "spotifete.baseUrl")
	configuredPort := getOptionalInt(viperConfiguration, "spotifete.port")
	if configuredPort == nil {
		c.Port = 8410
	} else {
		c.Port = *configuredPort
	}

	c.ReleaseMode = getBool(viperConfiguration, "spotifete.releaseMode")
	logDirectory := getOptionalString(viperConfiguration, "spotifete.logDirectory")
	if logDirectory == nil {
		c.LogDirectory = "./logs"
	} else {
		c.LogDirectory = *logDirectory
	}

	c.AppConfiguration = appConfiguration{}.read(viperConfiguration)

	return c
}
