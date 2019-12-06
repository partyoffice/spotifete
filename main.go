package main

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/webapp"
	"github.com/getsentry/sentry-go"
)

func main() {
	defer database.Shutdown()

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.GetConfig().GetString("sentry.dsn"),
	}); err != nil {
		panic("Sentry initialization failed: " + err.Error())
	}

	activeProfile := config.GetActiveProfile()
	webapp.Start(activeProfile)
}
