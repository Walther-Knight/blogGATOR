package main

import (
	"fmt"
	"os"

	"github.com/Walther-Knight/blogGATOR/internal/config"
)

func main() {
	// Read initial config
	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	// Set user
	if err := cfg.SetUser("brent"); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting user: %v\n", err)
		os.Exit(1)
	}

	// Read config again to verify changes
	updatedCfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading updated config: %v\n", err)
		os.Exit(1)
	}

	// Print the config
	fmt.Printf("%+v\n", updatedCfg)
}
