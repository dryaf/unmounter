package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kardianos/service"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	http.HandleFunc("/", basicAuth(listMounts))
	http.HandleFunc("/restart-autofs", basicAuth(restartAutofs))
	fmt.Println("Server started at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error(err)
	}
}

func (p *program) Stop(s service.Service) error {
	// Any clean-up or resource release logic here.
	return nil
}

// Define command-line flags or arguments for service management.
// For simplicity, this example checks os.Args directly.
func handleServiceControl(s service.Service) {
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
