package database

import (
	"fmt"
	"github.com/47-11/spotifete/config"
	. "github.com/47-11/spotifete/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var connectionUrl string
var Connection *gorm.DB

func Shutdown() {
	if Connection != nil {
		Connection.Close()
	}
}

func init() {
	c := config.GetConfig()
	connectionUrl = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", c.GetString("database.host"), c.GetString("database.port"), c.GetString("database.name"), c.GetString("database.user"), c.GetString("database.password"))
	// Automatically migrate the schema during startup
	db, err := gorm.Open("postgres", connectionUrl)
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Session{})

	Connection = db
}
