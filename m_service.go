package main

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
)

var serviceConfig = &service.Config{
	Name:        "unmounter",
	DisplayName: "unmounter",
	Description: "A web service to list and unmount devices.",
	UserName:    "unmounter",
	EnvVars: map[string]string{
		EnvVarAuthUser: username,
		EnvVarAuthPass: password,
	},
}

type systemService struct{}

func (p *systemService) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *systemService) run() {
	runWebServer()
}

func (p *systemService) Stop(s service.Service) error {
	// Any clean-up or resource release logic here.
	return nil
}

func handleServiceArgs(s service.Service) {
	if len(os.Args) < 2 {
		fmt.Println("Usage: myservice <command>")
		fmt.Println("Commands: install, uninstall, start, stop, restart")
		return
	}
	cmd := os.Args[1]
	switch cmd {
	case "install":
		err := s.Install()
		if err != nil {
			fmt.Println("Failed to install:", err)
			return
		}
		fmt.Println("Service installed")
	case "uninstall":
		err := s.Uninstall()
		if err != nil {
			fmt.Println("Failed to uninstall:", err)
			return
		}
		fmt.Println("Service uninstalled")
	case "start":
		err := s.Start()
		if err != nil {
			fmt.Println("Failed to start:", err)
			return
		}
		fmt.Println("Service started")
	case "stop":
		err := s.Stop()
		if err != nil {
			fmt.Println("Failed to stop:", err)
			return
		}
		fmt.Println("Service stopped")
	case "restart":
		err := s.Restart()
		if err != nil {
			fmt.Println("Failed to restart:", err)
			return
		}
		fmt.Println("Service restarted")
	default:
		fmt.Println("Invalid command")
		fmt.Println("Usage: myservice <command>")
		fmt.Println("Commands: install, uninstall, start, stop, restart")
	}
}
