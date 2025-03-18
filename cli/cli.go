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

	switch os.Args[1] {
	case "run":
		return run()
	case "ps":
		return c.ps()
	case "stop":
		return c.stop()
	case "child":
		// return child()
		return fmt.Errorf("command not implemented")
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}
