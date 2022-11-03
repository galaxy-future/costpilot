package main

import (
	"log"
	"net/http"
	"os"

	"github.com/galaxy-future/costpilot/tools"
	"github.com/gin-gonic/gin"
)

func output() error {
	if os.Getenv("ENV") == "docker" {
		return _runServer()
	}
	return _runCmd()
}

func _runCmd() error {
	if err := tools.ShowHtml("website/index.html"); err != nil {
		return err
	}
	log.Printf("I! costpilot analysis completed! (if the system browser does not open automatically, you can open website/index.html manually in brower)")
	return nil
}

func _runServer() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.StaticFS("/website", http.Dir("./website"))
	log.Println("visit http://localhost:8504/website , check the cost analysis")
	if err := r.Run(":8504"); err != nil {
		log.Printf("E! %v\n", err)
		return err
	}
	return nil
}
