package database

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
)

var connectionUrl string
var Connection *gorm.DB

func Shutdown() {
	if Connection != nil {
		_ = Connection.Close()
	}
}

func init() {
	c := config.GetConfig()
	connectionUrl = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", c.GetString("database.host"), c.GetString("database.port"), c.GetString("database.name"), c.GetString("database.user"), c.GetString("database.password"))

	db, err := gorm.Open("postgres", connectionUrl)
	if err  != nil {
		panic("failed to connect to database: " + err.Error())
	}

	// Run migrations
	log.Println("Connection aquired. Running database migrations")
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

	err = m.Up()
	// TODO: There probably is a way to do this properly. But this works for now
	if err != nil && "no change" != err.Error() {
		panic("could not execute migration: " + err.Error())
	}

	Connection = db
}
