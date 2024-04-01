package main

type ServiceStatus struct {
	Name   string
	Active bool
	Detail string
}

type Usage struct {
	Command string `json:"command"`
	PID     int    `json:"pid"`
	User    string `json:"user"`
	Name    string `json:"name"`
}

type Mount struct {
	Device     string  `json:"device"`
	Path       string  `json:"path"`
	Usages     []Usage `json:"usages"`
	UsageError string  `json:"usageError,omitempty"` // Holds error messages related to usage fetching
}

type SystemStatusResponse struct {
	Mounts []Mount       `json:"mounts"`
	Error  string        `json:"error,omitempty"`
	Autofs ServiceStatus `json:"autofs"`
	Samba  ServiceStatus `json:"samba"`
}
