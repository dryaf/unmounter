package main

import (
	"log"
	"os"

	"github.com/kardianos/service"
)

var logger service.Logger

// Configuration for the basic HTTP authentication
const EnvAuthUser = "AUTH_USER"
const EnvAuthPass = "AUTH_PASS"

var username = "admin"
var password = "sAcure1passw"

func init() {
	envUsername, ok := os.LookupEnv(EnvAuthUser)
	if ok {
		username = envUsername
	}

	envPassword, ok := os.LookupEnv(EnvAuthPass)
	if ok {
		password = envPassword
	}
}

func main() {
	svcConfig := &service.Config{
		Name:        "unmounter",
		DisplayName: "unmounter",
		Description: "A web service to list and unmount devices.",
		UserName:    "unmounter",
		EnvVars: map[string]string{
			EnvAuthUser: username,
			EnvAuthPass: password,
		},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		handleServiceControl(s)
		return
	}

	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if err = s.Run(); err != nil {
		logger.Error(err)
	}
}
