package main

import (
	"log"

	"github.com/galayx-future/costpilot/version"
)

var Version = version.Version

func printVersion() {
	log.Printf("[Galaxy-Future] CostPilot %v\n", Version)
}
