package config

import (
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

var instance Configuration
var once sync.Once

func Get() Configuration {
	once.Do(func() {
		instance = readConfiguration()
	})
	return instance
}

func readConfiguration() Configuration {
	viperConfiguration := readViperConfiguration()
	return Configuration{}.read(viperConfiguration)
}

func readViperConfiguration() *viper.Viper {
	viperConfiguration := createViperConfiguration()
	err := viperConfiguration.ReadInConfig()
	if err != nil {
		logger.Fatal("Could not read config file.")
	}

	return viperConfiguration
}

func createViperConfiguration() *viper.Viper {
	viperConfiguration := viper.New()
	viperConfiguration.SetConfigType("yaml")
	viperConfiguration.SetConfigName("spotifete-config")
	viperConfiguration.AddConfigPath("/etc/spotifete")
	viperConfiguration.AddConfigPath(".")
	return viperConfiguration;
}

func (c Configuration) read(viperConfiguration *viper.Viper) Configuration {
	c.SpotifeteConfiguration = spotifeteConfiguration{}.read(viperConfiguration)
	c.DatabaseConfiguration = databaseConfiguration{}.read(viperConfiguration)
	c.SpotifyConfiguration = spotifyConfiguration{}.read(viperConfiguration)
	c.SentryConfiguration = sentryConfiguration{}.read(viperConfiguration)

	return c
}

func getRequiredString(viperConfiguration *viper.Viper, key string) string {
	if viperConfiguration.IsSet(key) {
		return viperConfiguration.GetString(key)
	} else {
		logger.Fatalf("Required string configuration parameter %s is not present.", key)
		panic("Incomplete configuration")
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
		logger.Fatalf("Required int configuration parameter %s is not present.", key)
		panic("Incomplete configuration")
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
