package config

import (
	"fmt"
	"github.com/google/logger"
	"github.com/spf13/viper"
	"sync"
)

type Configuration struct {
	SpotifeteConfiguration spotifeteConfiguration
	DatabaseConfiguration  databaseConfiguration
	SpotifyConfiguration   spotifyConfiguration
	SentryConfiguration    sentryConfiguration
}

type spotifeteConfiguration struct {
	BaseUrl          string
	ReleaseMode      bool
	AppConfiguration appConfiguration
}

type appConfiguration struct {
	AndroidUrl *string
	IOsUrl     *string
}

type databaseConfiguration struct {
	Host       string
	Port       int
	Name       string
	User       string
	Password   string
	DisableSsl bool
}

type spotifyConfiguration struct {
	Id     string
	Secret string
}

type sentryConfiguration struct {
	Dsn *string
}

var instance Configuration
var once sync.Once

func Get() Configuration {
	once.Do(func() {
		viperConfig := viper.New()
		viperConfig.SetConfigType("yaml")
		viperConfig.SetConfigName("spotifete-config")
		viperConfig.AddConfigPath("/etc/spotifete")
		viperConfig.AddConfigPath(".")
		err := viperConfig.ReadInConfig()
		if err != nil {
			logger.Fatal("Could not read config file.")
		}

		instance = Configuration{}.read(viperConfig)
	})
	return instance
}

func getRequiredString(viperConfiguration *viper.Viper, key string) string {
	if viperConfiguration.IsSet(key) {
		return viperConfiguration.GetString(key)
	} else {
		panic(fmt.Sprintf("Required string configuration parameter %s is not present.", key))
	}
}

func getOptionalString(viperConfiguration *viper.Viper, key string) *string {
	if viperConfiguration.IsSet(key) {
		value := viperConfiguration.GetString(key)
		return &value
	} else {
		return nil
	}
}

func getRequiredInt(viperConfiguration *viper.Viper, key string) int {
	if viperConfiguration.IsSet(key) {
		return viperConfiguration.GetInt(key)
	} else {
		panic(fmt.Sprintf("Required int configuration parameter %s is not present.", key))
	}
}

func getOptionalInt(viperConfiguration *viper.Viper, key string) *int {
	if viperConfiguration.IsSet(key) {
		value := viperConfiguration.GetInt(key)
		return &value
	} else {
		return nil
	}
}

func getBool(viperConfiguration *viper.Viper, key string) bool {
	if viperConfiguration.IsSet(key) {
		return viperConfiguration.GetBool(key)
	} else {
		logger.Warningf("Bool configuration parameter %s not present, falling back to default false. Explicitly set a value to disable this warning.", key)
		return false
	}
}

func (c Configuration) read(viperConfiguration *viper.Viper) Configuration {
	c.SpotifeteConfiguration = spotifeteConfiguration{}.read(viperConfiguration)
	c.DatabaseConfiguration = databaseConfiguration{}.read(viperConfiguration)
	c.SpotifyConfiguration = spotifyConfiguration{}.read(viperConfiguration)
	c.SentryConfiguration = sentryConfiguration{}.read(viperConfiguration)

	return c
}

func (c spotifeteConfiguration) read(viperConfiguration *viper.Viper) spotifeteConfiguration {
	c.BaseUrl = getRequiredString(viperConfiguration, "spotifete.baseUrl")
	c.ReleaseMode = getBool(viperConfiguration, "spotifete.releaseMode")
	c.AppConfiguration = appConfiguration{}.read(viperConfiguration)

	return c
}

func (c appConfiguration) read(viperConfiguration *viper.Viper) appConfiguration {
	c.AndroidUrl = getOptionalString(viperConfiguration, "spotifete.app.androidUrl")
	c.IOsUrl = getOptionalString(viperConfiguration, "spotifete.app.iosUrl")

	return c
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

func (c databaseConfiguration) GetConnectionUrl() string {
	connectionUrl := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s", c.Host, c.Port, c.Name, c.User, c.Password)
	if c.DisableSsl {
		connectionUrl += " sslmode=disable"
	}

	return connectionUrl
}

func (c spotifyConfiguration) read(viperConfiguration *viper.Viper) spotifyConfiguration {
	c.Id = getRequiredString(viperConfiguration, "spotify.id")
	c.Secret = getRequiredString(viperConfiguration, "spotify.secret")

	return c
}

func (c sentryConfiguration) read(viperConfiguration *viper.Viper) sentryConfiguration {
	c.Dsn = getOptionalString(viperConfiguration, "sentry.dsn")

	return c
}
