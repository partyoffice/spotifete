package database

import (
	"github.com/47-11/spotifete/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"sync"
)

const targetDatabaseVersion = 30

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

func migrateIfNecessary(db *gorm.DB) {
	logger.Info("Connection aquired. Checking database version")
	migration := prepareMigration(db)
	runMigrationIfNecessary(migration)
}

func prepareMigration(db *gorm.DB) *migrate.Migrate {
	driver := getDatabaseDriver(db)

	migration, err := migrate.NewWithDatabaseInstance(
		"file://resources/migrations/",
		"postgres", driver)
	if err != nil {
		logger.Fatalf("could not prepare database migration: %s", err.Error())
	}

	return migration
}

func getDatabaseDriver(db *gorm.DB) database.Driver {
	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
	if err != nil {
		logger.Fatalf("could not get driver for migration from db instance: %s", err.Error())
	}

	return driver
}

func runMigrationIfNecessary(migration *migrate.Migrate) {
	version, _, _ := migration.Version()
	if version == targetDatabaseVersion {
		logger.Infof("Database is up to date! (Version %d)", version)
	} else {
		logger.Infof("Database version is %d / target version is %d. Migrating!\n", version, targetDatabaseVersion)
		runMigration(migration)
	}
}

func runMigration(migration *migrate.Migrate) {
	err := migration.Migrate(targetDatabaseVersion)
	if err != nil {
		logger.Fatalf("could not execute migration: %s", err.Error())
	}

	logger.Info("Migration successful!")
}
