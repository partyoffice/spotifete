package database

import (
	"github.com/47-11/spotifete/config"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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

func CloseConnection() {
	logger.Info("Closing db connection")
	if connection != nil {
		err := connection.Close()
		logger.Error(err)
	}
}

func initialize() *gorm.DB {
	db := openConnection()
	migrateIfNecessary(db)
	return db
}

func openConnection() *gorm.DB {
	db, err := gorm.Open("postgres", config.Get().DatabaseConfiguration.BuildConnectionUrl())
	if err != nil {
		logger.Fatalf("failed to connect to database: %s", err.Error())
	}

	return db
}
