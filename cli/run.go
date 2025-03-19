package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	// fs := container.SetupFilesystem()
	// if err := fs.Setup(); err != nil {
	// 	return fmt.Errorf("failed to set up filesystem: %v", err)
	// }

	// Set up network if needed
	// network := container.SetupNetwork()
	// if err := network.Setup(); err != nil {
	// 	return fmt.Errorf("failed to set up network: %v", err)
	// }

	// Start the container
	if err := container.Start(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	fmt.Printf("Container started with ID: %s\n", container.ID)
	return nil
}

func runChild() error {
	fmt.Printf("Running %v \n", os.Args[2:])
	// cg()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// must(syscall.Sethostname([]byte("container")))
	// must(syscall.Chroot("/home/beltran/ubuntufs"))
	// must(os.Chdir("/"))
	// must(syscall.Mount("proc", "proc", "proc", 0, ""))
	// must(syscall.Mount("thing", "mytemp", "tempfs", 0, ""))

	must(cmd.Run())

	// must(syscall.Unmount("proc", 0))
	// must(syscall.Unmount("thing", 0))

	return nil
}

func cg() {
	cgroups := "/sys/fs/cgroup"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "beltran"), 0755)
	must(os.WriteFile(filepath.Join(pids, "beltran", "pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	must(os.WriteFile(filepath.Join(pids, "beltran", "notify_on_release"), []byte("1"), 0700))
	must(os.WriteFile(filepath.Join(pids, "beltran", "cgroup.procs"), []byte(fmt.Sprintf("%d", os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
