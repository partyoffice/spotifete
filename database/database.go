package database

import (
	"github.com/47-11/spotifete/config"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
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
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: config.Get().DatabaseConfiguration.BuildConnectionUrl(),
	}), &gorm.Config{})

	if err != nil {
		logger.Fatalf("failed to connect to database: %s", err.Error())
	}

	return db
}
