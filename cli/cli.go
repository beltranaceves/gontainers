package cli

import (
	"fmt"
	"os"
)

type CLI struct {
}

type containerInfo struct {
	id      string
	command string
	pid     string
	status  string
}

func NewCLI() *CLI {
	return &CLI{}
}

func (c *CLI) Execute() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("command required")
	}

	fmt.Printf("Running %v as user %d in process %d \n", os.Args[1:], os.Getuid(), os.Getpid())

	switch os.Args[1] {
	case "run":
		return runParent()
	case "ps":
		return ps()
	case "stop":
		return stop()
	case "child":
		return runChild()
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}
