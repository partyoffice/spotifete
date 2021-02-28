package config

import (
	"github.com/spf13/viper"
	"strings"
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
	c.LogDirectory = getRequiredString(viperConfiguration, "spotifete.logDirectory")
	if !strings.HasSuffix(c.LogDirectory, "/") {
		c.LogDirectory = c.LogDirectory + "/"
	}

	c.AppConfiguration = appConfiguration{}.read(viperConfiguration)

	return c
}
