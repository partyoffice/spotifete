package database

import (
	"fmt"
	. "github.com/47-11/spotifete/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var instance *gorm.DB

func GetInstance() *gorm.DB {
	if instance == nil {
		db, err := gorm.Open("postgres", "host=nikos410.de port=5432 user=spotifete dbname=spotifete password=?")
		if err != nil {
			panic("failed to connect database")
		}

		instance = db
	}

	return instance
}

func Shutdown() {
	if instance != nil {
		fmt.Println("Closing database")
		instance.Close()
	}
}

func init() {
	// Automatically migrate the schema during startup
	db, err := gorm.Open("postgres", "host=nikos410.de port=5432 user=spotifete dbname=spotifete password=?")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	fmt.Println("Initializing database")

	// Migrate the schema
	db.AutoMigrate(&Session{})
}
