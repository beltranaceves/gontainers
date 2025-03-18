package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/beltranaceves/gontainers/container"
)

func runParent() error {
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

func runChild() error {
	fmt.Printf("Running %v \n", os.Args[2:])
	cg()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/home/beltran/ubuntufs"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	must(syscall.Mount("thing", "mytemp", "tempfs", 0, ""))

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
	must(syscall.Unmount("thing", 0))

	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
