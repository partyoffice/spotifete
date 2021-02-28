package main

import (
	"github.com/47-11/spotifete/database"
	"github.com/47-11/spotifete/listeningSession"
	"github.com/47-11/spotifete/logging"
	"github.com/47-11/spotifete/webapp"
	"io/ioutil"
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

	println(string(bannerTextBytes))
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
