package main

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/listeningSession"
	"github.com/47-11/spotifete/logging"
	"github.com/47-11/spotifete/webapp"
)

var spotifeteWebapp webapp.SpotifeteWebapp

func main() {
	setup()
	run()
}

func setup() {
	logging.SetupLogging()
	database.GetConnection()
	setupWebapp()
}

func setupWebapp() {
	spotifeteWebapp = webapp.SpotifeteWebapp{}.Setup()
}

func run() {
	listeningSession.StartPollSessionsLoop()
	spotifeteWebapp.Run()
}