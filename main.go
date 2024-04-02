package main

import (
	"log"
	"os"

	"github.com/kardianos/service"
)

var logger service.Logger

func main() {
	systemService, err := service.New(&systemService{}, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		handleServiceArgs(systemService)
		return
	}

	logger, err = systemService.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if err = systemService.Run(); err != nil {
		logger.Error(err)
	}
}
