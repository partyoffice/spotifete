package config

import "github.com/spf13/viper"

type spotifyConfiguration struct {
	Id     string
	Secret string
}

func (c spotifyConfiguration) read(viperConfiguration *viper.Viper) spotifyConfiguration {
	c.Id = getRequiredString(viperConfiguration, "spotify.id")
	c.Secret = getRequiredString(viperConfiguration, "spotify.secret")

	return c
}
