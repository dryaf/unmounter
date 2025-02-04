// ==== File: m_dev_mode.go ====
// ==== File: m_dev_mode.go ====
package main

import (
	"fmt"
	"strings"
	"time"
)

func checkAutofsStatusDevMode() ServiceStatus {
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
	return ServiceStatus{Name: "Autofs", Active: true, Detail: detail}
}

func checkSambaStatusDevMode() ServiceStatus {
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
258080       1001       DENY_NONE  0x120089    RDONLY     NONE             /mnt/external   audio/bob-says-hello.flac   Tue Feb  4 17:33:57 2025`
	return ServiceStatus{Name: "Samba", Active: false, Detail: detail} // Simulating locked files, so Active: false
}

func getMountsDevMode() []Mount {
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
					Name:    "/mnt/external/audio/bob-says-hello.flac", // Simulate the second locked file entry
				},
			},
			UsageError:          "",
			FreeSpace:           "2.5 GB",
			TotalSpace:          "10 GB", // Simulated total space
			UsedSpacePercentage: 75,      // Simulated used space percentage
			FreeSpacePercentage: 25,      // Simulated free space percentage
		},
		{
			Device:              "/dev/sdb2",
			Path:                "/media/usb0",
			Usages:              []Usage{}, // No usages for this one in dev mode sample
			UsageError:          "",
			FreeSpace:           "500 MB",
			TotalSpace:          "2 GB", // Simulated total space
			UsedSpacePercentage: 80,     // Simulated used space percentage
			FreeSpacePercentage: 20,     // Simulated free space percentage
		},
		{
			Device:              "/dev/sdc1",
			Path:                "/mnt/fail_unmount", // Simulate device that fails to unmount
			Usages:              []Usage{},
			UsageError:          "",
			FreeSpace:           "10 GB",
			TotalSpace:          "100 GB", // Simulated total space
			UsedSpacePercentage: 90,       // Simulated used space percentage
			FreeSpacePercentage: 10,       // Simulated free space percentage
		},
	}
	return mounts
}

func getUsagesDevMode(mountPoint string) ([]Usage, string) {
	time.Sleep(100 * time.Millisecond) // Simulate delay
	if strings.Contains(mountPoint, "usb0") {
		return []Usage{
			{Command: "mock_process1", PID: 1234, User: "mockuser", Name: "mock_file1.txt"},
			{Command: "mock_process2", PID: 5678, User: "mockuser", Name: "mock_file2.txt"},
		}, ""
	}
	return []Usage{}, ""
}

func unmountDeviceDevMode(device string) error {
	time.Sleep(200 * time.Millisecond)    // Simulate delay
	if strings.Contains(device, "fail") { // Simulate unmount failure for devices containing "fail"
		return fmt.Errorf("simulated unmount failure for device: %s", device)
	}
	return nil // Simulate successful unmount
}

func killProcessDevMode(pid int) error {
	time.Sleep(100 * time.Millisecond) // Simulate delay
	if pid == 9999 {                   // Simulate kill failure for PID 9999
		return fmt.Errorf("simulated kill process failure for pid: %d", pid)
	}
	return nil // Simulate successful kill
}

func getDiskFreeSpaceDevMode() (string, int, error) {
	time.Sleep(50 * time.Millisecond) // Simulate delay
	return "1.23 GB", 60, nil         // Simulated free space and percentage
}
