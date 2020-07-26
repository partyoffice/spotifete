package config

import "github.com/spf13/viper"

type sentryConfiguration struct {
	Dsn *string
}

func (c sentryConfiguration) read(viperConfiguration *viper.Viper) sentryConfiguration {
	c.Dsn = getOptionalString(viperConfiguration, "sentry.dsn")

	return c
}
