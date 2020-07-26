package config

import "github.com/spf13/viper"

type appConfiguration struct {
	AndroidUrl *string
	IOsUrl     *string
}

func (c appConfiguration) read(viperConfiguration *viper.Viper) appConfiguration {
	c.AndroidUrl = getOptionalString(viperConfiguration, "spotifete.app.androidUrl")
	c.IOsUrl = getOptionalString(viperConfiguration, "spotifete.app.iosUrl")

	return c
}
