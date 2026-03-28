//go:build !windows

package elevation

import "os"

// IsElevated returns true if the current process is running as root (uid 0).
func IsElevated() bool {
	return os.Geteuid() == 0
}

// relaunchElevated is a no-op on Unix; we just print a message and exit instead.
func relaunchElevated() error {
	return nil
}
