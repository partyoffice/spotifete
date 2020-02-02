package database

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"sync"
)

const targetDatabaseVersion = 29

var connectionUrl string
var connection *gorm.DB
var once sync.Once

func Shutdown() {
	if connection != nil {
		err := connection.Close()
		logger.Error(err)
	}
}

func GetConnection() *gorm.DB {
	once.Do(func() {
		initialize()
	})

	return connection
}

func initialize() {
	c := config.GetConfig()

	disableSsl := ""
	if c.GetBool("database.disableSsl") {
		disableSsl = " sslmode=disable"
	}
	connectionUrl = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s %s", c.GetString("database.host"), c.GetString("database.port"), c.GetString("database.name"), c.GetString("database.user"), c.GetString("database.password"), disableSsl)

	db, err := gorm.Open("postgres", connectionUrl)
	if err != nil {
		logger.Fatalf("failed to connect to database: %s", err.Error())
	}

	// Run migrations
	logger.Info("Connection aquired. Checking database version")
	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
	if err != nil {
		logger.Fatalf("could not get driver for migration from db instance: %s", err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://resources/migrations/",
		"postgres", driver)
	if err != nil {
		logger.Fatalf("could not prepare database migration: " + err.Error())
	}

	version, _, _ := m.Version()
	if version != targetDatabaseVersion {
		logger.Infof("Database version is %d / target version is %d. Migrating!\n", version, targetDatabaseVersion)
		err = m.Migrate(targetDatabaseVersion)
		if err != nil {
			logger.Fatalf("could not execute migration: %s", err.Error())
		}

		logger.Info("Migrations successful!")
	} else {
		logger.Infof("Database is up to date! (Version %d)", version)
	}

	connection = db
}
