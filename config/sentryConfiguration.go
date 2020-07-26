package config

import (
	"github.com/getsentry/sentry-go"
	"github.com/spf13/viper"
)

type sentryConfiguration struct {
	Dsn *string
}

func (c sentryConfiguration) read(viperConfiguration *viper.Viper) sentryConfiguration {
	c.Dsn = getOptionalString(viperConfiguration, "sentry.dsn")

	return c
}

func (c sentryConfiguration) GetSentryClientOptions() sentry.ClientOptions {
	return sentry.ClientOptions{
		Dsn:              *c.Dsn,
		AttachStacktrace: true,
		IgnoreErrors:     []string{".*Refresh token revoked.*"},
	}
}
