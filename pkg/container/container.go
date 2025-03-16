package container

import (
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
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (c *Container) Start() error {
	cmd := exec.Command(c.Command, c.Args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUTS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	c.Pid = cmd.Process.Pid // Store the PID
	return nil
}
func (c *Container) Kill() error {
	// Implementation to kill the container process
	return syscall.Kill(c.Pid, syscall.SIGTERM)
}
