package cli

import (
	"fmt"
	"os"

	"github.com/beltranaceves/gontainers/container"
)

func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("command required for run")
	}

	// Extract command and arguments
	command := os.Args[2]
	args := []string{}
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}

	// Create a new container
	container := container.NewContainer(command, args)

	// Set up filesystem
	fs := container.SetupFilesystem()
	if err := fs.Setup(); err != nil {
		return fmt.Errorf("failed to set up filesystem: %v", err)
	}

	// Set up network if needed
	network := container.SetupNetwork()
	if err := network.Setup(); err != nil {
		return fmt.Errorf("failed to set up network: %v", err)
	}

	// Start the container
	if err := container.Start(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	fmt.Printf("Container started with ID: %s\n", container.ID)
	return nil
}
