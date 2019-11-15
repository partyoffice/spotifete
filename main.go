package main

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/webapp"
)

func main() {
	defer database.Shutdown()

	activeProfile := config.GetActiveProfile()
	webapp.Start(activeProfile)
}
