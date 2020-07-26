package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type databaseConfiguration struct {
	Host       string
	Port       int
	Name       string
	User       string
	Password   string
	DisableSsl bool
}

func (c databaseConfiguration) read(viperConfiguration *viper.Viper) databaseConfiguration {
	c.Host = getRequiredString(viperConfiguration, "database.host")
	c.Port = getRequiredInt(viperConfiguration, "database.port")
	c.Name = getRequiredString(viperConfiguration, "database.name")
	c.User = getRequiredString(viperConfiguration, "database.user")
	c.Password = getRequiredString(viperConfiguration, "database.password")
	c.DisableSsl = getBool(viperConfiguration, "database.disableSsl")

	return c
}

func (c databaseConfiguration) BuildConnectionUrl() string {
	connectionUrl := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s", c.Host, c.Port, c.Name, c.User, c.Password)
	if c.DisableSsl {
		connectionUrl += " sslmode=disable"
	}

	return connectionUrl
}
