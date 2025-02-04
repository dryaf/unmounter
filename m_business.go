// ==== File: m_business.go ====
package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

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
	Device              string  `json:"device"`
	Path                string  `json:"path"`
	Usages              []Usage `json:"usages"`
	UsageError          string  `json:"usageError,omitempty"`
	FreeSpace           string  `json:"freeSpace,omitempty"`
	FreeSpacePercentage int     `json:"freeSpacePercentage,omitempty"`
}

type SystemStatus struct {
	Mounts []Mount       `json:"mounts"`
	AutoFs ServiceStatus `json:"autofs"`
	Samba  ServiceStatus `json:"samba"`

	ErrorMounts error
	ErrorAutoFs error
	ErrorSamba  error
}

func getSystemStatus() *SystemStatus {
	response := &SystemStatus{}
	response.Mounts, response.ErrorMounts = getMounts()
	response.AutoFs, response.ErrorAutoFs = checkAutofsStatus()
	response.Samba, response.ErrorSamba = checkSambaStatus()
	return response
}

func checkAutofsStatus() (ServiceStatus, error) {
	if devModeEnabled {
		devStatus := checkAutofsStatusDevMode() // Call dev-mode function
		return devStatus, nil
	}
	cmd := exec.Command("systemctl", "status", "autofs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ServiceStatus{}, err
	}

	active := strings.Contains(string(output), "active (running)")
	return ServiceStatus{Name: "Autofs", Active: active, Detail: strings.TrimSpace(string(output))}, nil
}

func checkSambaStatus() (ServiceStatus, error) {
	if devModeEnabled {
		devStatus := checkSambaStatusDevMode() // Call dev-mode function
		return devStatus, nil
	}
	cmd := exec.Command("sudo", "smbstatus", "--locked")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ServiceStatus{}, err
	}

	noLockedFiles := strings.Contains(string(output), "No locked files")
	return ServiceStatus{Name: "Samba", Active: noLockedFiles, Detail: string(output)}, nil
}

func unmountDevice(device string) error {
	if devModeEnabled {
		return unmountDeviceDevMode(device) // Call dev-mode function
	}

	// Check if the device is in the list of currently mounted devices
	mounts, err := getMounts()
	if err != nil {
		return fmt.Errorf("failed to get mounts: %v", err)
	}

	found := false
	for _, mount := range mounts {
		if device == mount.Path {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("device not mounted: %s", device)
	}

	// Device is valid and mounted, proceed with unmount
	cmd := exec.Command("sudo", "umount", "--", device)
	err = cmd.Run()
	if err != nil {
		// Convert error to *exec.ExitError and get the exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 1:
				return fmt.Errorf("failed to unmount device: %s, error: an unspecified error occurred", device)
			case 2:
				return fmt.Errorf("failed to unmount device: %s, error: permission denied (are you root?)", device)
			case 8:
				return fmt.Errorf("failed to unmount device: %s, error: no such file or directory", device)
			case 16:
				return fmt.Errorf("failed to unmount device: %s, error: device is busy", device)
			case 32:
				return fmt.Errorf("failed to unmount device: %s, error: device is busy - umount command failed", device)
			default:
				return fmt.Errorf("failed to unmount device: %s, error: unknown error with exit status %d", device, exitErr.ExitCode())
			}
		}
		// For non-exit errors, just return the error itself
		return fmt.Errorf("failed to unmount device: %s, error: %v", device, err)
	}

	return nil
}

func killProcess(pid int) error {
	if devModeEnabled {
		return killProcessDevMode(pid) // Call dev-mode function
	}
	mounts, err := getMounts()
	if err != nil {
		return fmt.Errorf("failed to get mounts: %v", err)
	}

	found := false
	for _, mount := range mounts {
		for _, usage := range mount.Usages {
			if pid == usage.PID {
				found = true
				break
			}
		}
	}
	if !found {
		return fmt.Errorf("pid not found: %d", pid)
	}

	return exec.Command("sudo", "kill", "-9", strconv.Itoa(pid)).Run()
}

func getDiskFreeSpace(path string) (string, int, error) {
	if devModeEnabled {
		return getDiskFreeSpaceDevMode() // Call dev-mode function
	}
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return "", 0, err
	}
	// Available blocks * size per block = available space in bytes
	freeBytes := stat.Bavail * uint64(stat.Bsize)
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	percentage := int(float64(freeBytes) / float64(totalBytes) * 100)
	return formatBytes(freeBytes), percentage, nil
}

// Helper function to format bytes into human-readable format (MB or GB)
func formatBytes(bytes uint64) string {
	const (
		MB = 1 << 20
		GB = 1 << 30
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func getMounts() ([]Mount, error) {
	if devModeEnabled {
		devMounts := getMountsDevMode() // Call dev-mode function
		return devMounts, nil
	}
	cmd := exec.Command("mount")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var mounts []Mount
	pattern := regexp.MustCompile(`(\/dev\/[^\s]+)\s+on\s+([^\s]+(?:\s+[^\s]+)*?)\s+type`)
	for _, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) == 3 {
			mountSource := matches[1]
			mountPoint := matches[2]
			if strings.HasPrefix(mountSource, "/dev/sd") && (strings.HasPrefix(mountPoint, "/mnt/") || strings.HasPrefix(mountPoint, "/media/")) {
				usages, usageError := getUsages(mountPoint)
				freeSpace, freeSpacePercentage, err := getDiskFreeSpace(mountPoint)
				if err != nil {
					freeSpace = "Error fetching free space"
					logger.Error("Error getting free space for", mountPoint, ":", err)
				}
				mount := Mount{
					Device:              mountSource,
					Path:                mountPoint,
					Usages:              usages,
					UsageError:          usageError,
					FreeSpace:           freeSpace,
					FreeSpacePercentage: freeSpacePercentage,
				}
				mounts = append(mounts, mount)
			}
		}
	}
	return mounts, nil
}

func getUsages(mountPoint string) ([]Usage, string) {
	if devModeEnabled {
		devUsages, devError := getUsagesDevMode(mountPoint) // Call dev-mode function
		return devUsages, devError
	}
	cmd := exec.Command("sudo", "lsof", "--", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return []Usage{}, "" // No usages and no error.
		}
		return nil, fmt.Sprintf("error executing lsof on %s: %v", mountPoint, err)
	}

	var usages []Usage
	lines := strings.Split(string(output), "\n")
	// Skip the header line and any empty lines.
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		// Regex to match each field, considering that the NAME field can contain spaces.
		pattern := regexp.MustCompile(`^(\S+)\s+(\d+)\s+(\S+)\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+(.*)$`)
		matches := pattern.FindStringSubmatch(line)
		if matches == nil || len(matches) != 5 {
			continue // Skip if the line doesn't match the pattern.
		}
		pid, convErr := strconv.Atoi(matches[2])
		if convErr != nil {
			continue // Skip if PID conversion fails.
		}
		usage := Usage{
			Command: matches[1],
			PID:     pid,
			User:    matches[3],
			Name:    matches[4],
		}
		usages = append(usages, usage)
	}

	return usages, "" // Return usages with no error.
}
