package database

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/logger"
	"github.com/jinzhu/gorm"
)

const targetDatabaseVersion = 30

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
	version, dirty, _ := migration.Version()
	if dirty {
		logger.Fatal("Could not run migration. Current database is dirty.")
	} else if version == targetDatabaseVersion {
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
