package utils

import (
	"bytes"
	"gRPCClient/models"

	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

// Fetch details of the host system.
func FetchComputerSystemDetails() (models.ComputerSystem, error) {

	// Placeholder variable.
	var computerSystem models.ComputerSystem

	computerSystemDetails := unix.Utsname{}
	err := unix.Uname(&computerSystemDetails)
	if err != nil {
		LogError("FetchOperatingSystemDetails", "error fetching operating system information", nil, err)
		return computerSystem, err
	}

	// Verify that valid data has been returned.
	if computerSystemDetails != (unix.Utsname{}) {

		// Set the currently logged in user.
		out, err := exec.Command("users").Output()
		if err != nil {
			LogError("FetchOperatingSystemDetails", "error executing the command", "users", err)
			computerSystem.CurrentLoggedInUser = ""
		} else {
			computerSystem.CurrentLoggedInUser = strings.TrimSpace(string(bytes.Trim(out, "\x00")))
		}

		computerSystem.Domain = strings.TrimSpace(strings.Replace(string(computerSystemDetails.Domainname[:]), "\x00", "", -1))
		if strings.Contains(computerSystem.Domain, "none") {
			computerSystem.Domain = ""
		}
		computerSystem.Hostname = strings.TrimSpace(strings.Replace(string(computerSystemDetails.Nodename[:]), "\x00", "", -1))
	}

	return computerSystem, err
}
