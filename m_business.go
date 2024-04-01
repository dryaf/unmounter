package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Define a struct for caching mounts
type mountCache struct {
	mounts     []Mount
	lastUpdate time.Time
	lock       sync.Mutex
}

var cache = mountCache{}
var ttl = time.Second * 5

func getSystemStatus() SystemStatusResponse {
	var response SystemStatusResponse

	mounts, err := getMounts()
	if err != nil {
		response.Error = "Failed to get mounts: " + err.Error()
	} else {
		response.Mounts = mounts
	}

	autofsStatus, err := checkAutofsStatus()
	if err != nil {
		logger.Error("Failed to get Autofs status:", err)
	} else {
		response.Autofs = autofsStatus
	}

	sambaStatus, err := checkSambaStatus()
	if err != nil {
		logger.Error("Failed to get Samba status:", err)
	} else {
		response.Samba = sambaStatus
	}

	return response
}

func getMounts() ([]Mount, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if time.Since(cache.lastUpdate) < ttl {
		return cache.mounts, nil
	}

	cmd := exec.Command("mount")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var mounts []Mount
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 2 {
			mountSource := parts[0] // The device or path being mounted
			mountPoint := parts[2]  // The target path where the device is mounted
			// Only proceed if mount source starts with `/dev/sd` and mount point starts with `/mnt/` or `/media/`
			if strings.HasPrefix(mountSource, "/dev/sd") && (strings.HasPrefix(mountPoint, "/mnt/") || strings.HasPrefix(mountPoint, "/media/")) && !strings.Contains(mountPoint, " ") {
				usages, usageError := getUsages(mountPoint)
				mount := Mount{
					Device:     mountSource,
					Path:       mountPoint,
					Usages:     usages,
					UsageError: usageError,
				}
				mounts = append(mounts, mount)
			}
		}
	}

	cache.mounts = mounts
	cache.lastUpdate = time.Now()

	return mounts, nil
}

func checkAutofsStatus() (ServiceStatus, error) {
	cmd := exec.Command("systemctl", "status", "autofs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ServiceStatus{}, err
	}

	active := strings.Contains(string(output), "active (running)")
	return ServiceStatus{Name: "Autofs", Active: active, Detail: strings.TrimSpace(string(output))}, nil
}

func checkSambaStatus() (ServiceStatus, error) {
	cmd := exec.Command("sudo", "smbstatus", "--locked")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ServiceStatus{}, err
	}

	noLockedFiles := strings.Contains(string(output), "No locked files")
	return ServiceStatus{Name: "Samba", Active: noLockedFiles, Detail: string(output)}, nil
}

func unmountDevice(device string) error {
	// First, ensure the device path is valid
	pattern := `^/(mnt|media)/[a-zA-Z0-9_-]+$`
	matched, err := regexp.MatchString(pattern, device)
	if err != nil {
		return fmt.Errorf("regex match error: %v", err)
	}
	if !matched {
		return fmt.Errorf("invalid device path: %s", device)
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
	cache = mountCache{}
	return nil
}

func restartAutofs(w http.ResponseWriter, r *http.Request) {
	// Make sure this is a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Execute the command to restart autofs
	cmd := exec.Command("sudo", "systemctl", "restart", "autofs")
	err := cmd.Run()
	if err != nil {
		logger.Error("Failed to restart autofs:", err)
		http.Error(w, "Failed to restart autofs", http.StatusInternalServerError)
		return
	}
	time.Sleep(2 * time.Second)
	cache = mountCache{}
	// Redirect to the main page, or just inform of success
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getUsages(mountPoint string) ([]Usage, string) {
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
	for _, line := range lines[1:] { // Skip the header line
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue // Skip if line doesn't have enough fields.
		}
		pid, convErr := strconv.Atoi(fields[1])
		if convErr != nil {
			continue // Skip if PID conversion fails.
		}
		usage := Usage{
			Command: fields[0],
			PID:     pid,
			User:    fields[2],
			Name:    fields[len(fields)-1],
		}
		usages = append(usages, usage)
	}

	return usages, "" // Return usages with no error.
}
