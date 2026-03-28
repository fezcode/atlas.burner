package main

import (
	"fmt"
	"os"

	"github.com/fezcode/atlas.burner/internal/elevation"
	"github.com/fezcode/atlas.burner/internal/tui"
)

var Version = "dev"

func main() {
	elevation.EnsureElevated()

	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
