package main

import (
	"github.com/47-11/spotifete/config"
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/service"
	"github.com/47-11/spotifete/webapp"
	"github.com/getsentry/sentry-go"
)

func main() {
	defer database.Shutdown()

	dsn := config.GetConfig().GetString("sentry.dsn")
	if len(dsn) > 0 {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: dsn,
		}); err != nil {
			panic("Sentry initialization failed: " + err.Error())
		}
	}

	go service.ListeningSessionService().PollSessions()

	activeProfile := config.GetActiveProfile()
	webapp.Start(activeProfile)
}
