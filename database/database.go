package database

import (
	"github.com/47-11/spotifete/config"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

var connection *gorm.DB
var once sync.Once

func GetConnection() *gorm.DB {
	once.Do(func() {
		connection = initialize()
	})

	return connection
}

func initialize() *gorm.DB {
	db := openConnection()
	migrateIfNecessary(db)
	return db
}

func openConnection() *gorm.DB {
	db, err := gorm.Open(buildDialector(), buildConfig())

	if err != nil {
		logger.Fatalf("failed to connect to database: %s", err.Error())
	}

	return db
}

func buildDialector() gorm.Dialector {
	return postgres.New(postgres.Config{
		DSN: config.Get().DatabaseConfiguration.BuildConnectionUrl(),
	})
}

func buildConfig() *gorm.Config {
	return &gorm.Config{
		Logger: buildLogger(),
	}
}

func buildLogger() gormLogger.Interface {
	var logLevel gormLogger.LogLevel
	if config.Get().SpotifeteConfiguration.ReleaseMode {
		logLevel = gormLogger.Warn
	} else {
		logLevel = gormLogger.Info
	}

	return gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormLogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logLevel,
			Colorful:      false,
		},
	)
}
