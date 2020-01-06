package database

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
)

const targetDatabaseVersion = 25

var connectionUrl string
var Connection *gorm.DB

func Shutdown() {
	if Connection != nil {
		_ = Connection.Close()
	}
}

func Initialize() {
	c := config.GetConfig()

	disableSsl := ""
	if c.GetBool("database.disableSsl") {
		disableSsl = " sslmode=disable"
	}
	connectionUrl = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s %s", c.GetString("database.host"), c.GetString("database.port"), c.GetString("database.name"), c.GetString("database.user"), c.GetString("database.password"), disableSsl)

	db, err := gorm.Open("postgres", connectionUrl)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	// Run migrations
	log.Println("Connection aquired. Checking database version")
	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
	if err != nil {
		panic("could not get driver for migration from db instance: " + err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://resources/migrations/",
		"postgres", driver)
	if err != nil {
		panic("could not prepare database migration: " + err.Error())
	}

	version, _, _ := m.Version()
	if version != targetDatabaseVersion {
		log.Printf("Database version is %d / target version is %d. Migrating!\n", version, targetDatabaseVersion)
		err = m.Migrate(targetDatabaseVersion)
		if err != nil {
			panic("could not execute migration: " + err.Error())
		}

		log.Println("Migrations successful!")
	} else {
		log.Printf("Database is up to date! (Version %d)", version)
	}

	Connection = db
}
