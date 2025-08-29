package main

import (
	"os"

	"github.com/google/dpi-accelerator/beckn-onix/internal/onixctl"
)

func main() {
	if err := onixctl.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}



