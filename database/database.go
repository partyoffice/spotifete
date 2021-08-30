package database

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/logger"
	"github.com/partyoffice/spotifete/config"
	"github.com/partyoffice/spotifete/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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
	c := config.Get()

	var logLevel gormLogger.LogLevel
	var logWriter io.Writer
	if c.SpotifeteConfiguration.ReleaseMode {
		logLevel = gormLogger.Error
		logWriter = io.MultiWriter(logging.OpenLogFile("gorm.log"), os.Stderr)
	} else {
		logLevel = gormLogger.Info
		logWriter = os.Stdout
	}

	return gormLogger.New(
		log.New(logWriter, "\n", log.LstdFlags),
		gormLogger.Config{
			SlowThreshold: time.Millisecond * 200,
			LogLevel:      logLevel,
			Colorful:      false,
		},
	)
}
