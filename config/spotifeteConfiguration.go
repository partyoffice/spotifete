package config

import "github.com/spf13/viper"

type spotifeteConfiguration struct {
	BaseUrl          string
	ReleaseMode      bool
	AppConfiguration appConfiguration
}

func (c spotifeteConfiguration) read(viperConfiguration *viper.Viper) spotifeteConfiguration {
	c.BaseUrl = getRequiredString(viperConfiguration, "spotifete.baseUrl")
	c.ReleaseMode = getBool(viperConfiguration, "spotifete.releaseMode")
	c.AppConfiguration = appConfiguration{}.read(viperConfiguration)

	return c
}
