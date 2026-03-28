package elevation

import (
	"fmt"
	"os"
	"runtime"
)

// EnsureElevated checks if the process has admin/root privileges.
// On Windows, it re-launches itself with UAC elevation if needed.
// On Unix, it prints a message and exits.
func EnsureElevated() {
	if IsElevated() {
		return
	}

	if runtime.GOOS == "windows" {
		err := relaunchElevated()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to elevate: %v\n", err)
			fmt.Fprintln(os.Stderr, "Please right-click the executable and select 'Run as administrator'.")
			os.Exit(1)
		}
		// The elevated process is now running; this one exits.
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "atlas.burner requires root privileges to write to USB devices.")
	fmt.Fprintln(os.Stderr, "Please re-run with: sudo "+os.Args[0])
	os.Exit(1)
}
