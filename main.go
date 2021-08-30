package main

import (
	"fmt"
	"io/ioutil"

	"github.com/partyoffice/spotifete/database"
	"github.com/partyoffice/spotifete/listeningSession"
	"github.com/partyoffice/spotifete/logging"
	"github.com/partyoffice/spotifete/webapp"
)

var spotifeteWebapp webapp.SpotifeteWebapp

func main() {
	printBanner()
	setup()
	run()
}

func printBanner() {
	bannerTextBytes, err := ioutil.ReadFile("resources/banner.txt")
	if err != nil {
		println("Could not read banner text file: " + err.Error())
		return
	}

	fmt.Println(string(bannerTextBytes))
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
