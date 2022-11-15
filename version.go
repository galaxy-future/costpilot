package main

import (
	"log"

	"github.com/galaxy-future/costpilot/version"
)

var Version = version.Version

func printVersion() {
	log.Printf("[Galaxy-Future] CostPilot %v\n", Version)
}
