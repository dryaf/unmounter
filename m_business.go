// ==== File: m_business.go ====
// ==== File: m_business.go ====
package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
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
		time.Sleep(100 * time.Millisecond) // Simulate delay
		detail := `● autofs.service - Automounts filesystems on demand
     Loaded: loaded (/lib/systemd/system/autofs.service; enabled; vendor preset: enabled)
     Active: active (running) since Sun 2025-01-26 21:36:00 CET; 1 weeks 1 days ago
       Docs: man:autofs(8)
   Main PID: 603 (automount)
      Tasks: 4 (limit: 3930)
        CPU: 56.020s
     CGroup: /system.slice/autofs.service
             └─603 /usr/sbin/automount --pid-file /var/run/autofs.pid`
		return ServiceStatus{Name: "Autofs", Active: true, Detail: detail}, nil
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
		time.Sleep(100 * time.Millisecond) // Simulate delay
		detail := `Samba version 4.13.13-Debian
PID     Username     Group        Machine                                   Protocol Version  Encryption           Signing
----------------------------------------------------------------------------------------------------------------------------------------
258080  sambauser    sambauser    192.168.4.107 (ipv4:192.168.4.107:52682)  SMB3_11           -                    partial(AES-128-CMAC)

Service      pid     Machine       Connected at                     Encryption   Signing
---------------------------------------------------------------------------------------------
ExternalDrive 258080  192.168.4.107 Tue Feb  4 16:03:32 2025 CET     -            -

Locked files:
Pid          User(ID)   DenyMode   Access      R/W        Oplock           SharePath   Name   Time
--------------------------------------------------------------------------------------------------
258080       1001       DENY_NONE  0x120089    RDONLY     NONE             /mnt/external   audio/Bob Dylan Playlist The Very Best Of Bob Dylan [FLAC] CD - Q/Bob_Dylan-Playlist_The_Very_Best_Of_Bob_Dylan-CD-FLAC-2014-FLACME/12-bob_dylan-things_have_changed.flac   Tue Feb  4 17:33:57 2025`
		return ServiceStatus{Name: "Samba", Active: false, Detail: detail}, nil // Simulating locked files, so Active: false
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
		time.Sleep(200 * time.Millisecond)    // Simulate delay
		if strings.Contains(device, "fail") { // Simulate unmount failure for devices containing "fail"
			return fmt.Errorf("simulated unmount failure for device: %s", device)
		}
		return nil // Simulate successful unmount
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
		time.Sleep(100 * time.Millisecond) // Simulate delay
		if pid == 9999 {                   // Simulate kill failure for PID 9999
			return fmt.Errorf("simulated kill process failure for pid: %d", pid)
		}
		return nil // Simulate successful kill
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
		time.Sleep(50 * time.Millisecond) // Simulate delay
		return "1.23 GB", 60, nil         // Simulated free space and percentage
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
		time.Sleep(150 * time.Millisecond) // Simulate delay
		mounts := []Mount{
			{
				Device: "/dev/sda1",
				Path:   "/mnt/external",
				Usages: []Usage{
					{
						Command: "smbd",
						PID:     258080,
						User:    "sambauser",
						Name:    "/mnt/external", // Simulate the first locked file entry
					},
					{
						Command: "smbd",
						PID:     258080,
						User:    "sambauser",
						Name:    "/mnt/external/audio/Bob Dylan Playlist The Very Best Of Bob Dylan [FLAC] CD - Q/Bob_Dylan-Playlist_The_Very_Best_Of_Bob_Dylan-CD-FLAC-2014-FLACME/12-bob_dylan-things_have_changed.flac", // Simulate the second locked file entry
					},
				},
				UsageError:          "",
				FreeSpace:           "2.5 GB",
				FreeSpacePercentage: 70,
			},
			{
				Device:              "/dev/sdb2",
				Path:                "/media/usb0",
				Usages:              []Usage{}, // No usages for this one in dev mode sample
				UsageError:          "",
				FreeSpace:           "500 MB",
				FreeSpacePercentage: 25,
			},
			{
				Device:              "/dev/sdc1",
				Path:                "/mnt/fail_unmount", // Simulate device that fails to unmount
				Usages:              []Usage{},
				UsageError:          "",
				FreeSpace:           "10 GB",
				FreeSpacePercentage: 90,
			},
		}
		return mounts, nil
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
				// Corrected line: append a single 'mount' struct, not the 'mounts' slice
				mounts = append(mounts, mount)
			}
		}
	}
	return mounts, nil
}

func getUsages(mountPoint string) ([]Usage, string) {
	if devModeEnabled {
		time.Sleep(100 * time.Millisecond) // Simulate delay
		if strings.Contains(mountPoint, "usb0") {
			return []Usage{
				{Command: "mock_process1", PID: 1234, User: "mockuser", Name: "mock_file1.txt"},
				{Command: "mock_process2", PID: 5678, User: "mockuser", Name: "mock_file2.txt"},
			}, ""
		}
		return []Usage{}, ""
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
