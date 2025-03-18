package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Container struct {
	ID       string
	Command  string
	Args     []string
	RootFS   string
	Network  *Network
	Resource *ResourceConfig
	Pid      int
}

type ResourceConfig struct {
	Memory    int64
	CPUShare  int64
	CPUPeriod int64
}

func NewContainer(command string, args []string) *Container {
	return &Container{
		ID:      generateID(),
		Command: command,
		Args:    args,
		Resource: &ResourceConfig{
			Memory:    512 * 1024 * 1024, // 512MB default
			CPUShare:  1024,              // Default CPU share
			CPUPeriod: 100000,
		},
	}
}

func generateID() string {
	return fmt.Sprintf("gontainer-%d", time.Now().UnixNano())
}

func (c *Container) Start() error {
	cmd := exec.Command(c.Command, c.Args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUTS,
	}

	// Set the process name to include the container ID for easier identification
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL

	// Use the configured rootfs
	if c.RootFS != "" {
		// Set the root directory for the container
		cmd.Dir = c.RootFS

		// Optional: chroot to the rootfs (requires root privileges)
		// This would require additional code to handle chroot
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	c.Pid = cmd.Process.Pid // Store the PID

	// Store container info for later retrieval
	if err := c.saveContainerInfo(); err != nil {
		return fmt.Errorf("failed to save container info: %v", err)
	}

	return nil
}

func (c *Container) saveContainerInfo() error {
	// Create directory inside the project
	infoDir := "./containers"
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		return err
	}

	// Extract ID without the "gontainer-" prefix
	idWithoutPrefix := c.ID
	if len(c.ID) > 10 && c.ID[:10] == "gontainer-" {
		idWithoutPrefix = c.ID[10:]
	}

	// Save basic container info to a file
	infoPath := fmt.Sprintf("%s/%s.json", infoDir, idWithoutPrefix)
	info := map[string]interface{}{
		"id":      c.ID,
		"command": c.Command,
		"args":    c.Args,
		"pid":     c.Pid,
		"rootfs":  c.RootFS,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return os.WriteFile(infoPath, jsonData, 0644)
}

func (c *Container) SetupFilesystem() *Filesystem {
	// Create a unique root filesystem path for this container
	rootPath := fmt.Sprintf("./containers/%s/rootfs", c.ID[10:])
	fs := NewFilesystem(rootPath)
	c.RootFS = rootPath
	return fs
}
func (c *Container) Kill() error {
	// Implementation to kill the container process
	return syscall.Kill(c.Pid, syscall.SIGTERM)
}
