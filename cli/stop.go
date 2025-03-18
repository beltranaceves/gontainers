package cli

import (
	"fmt"
	"os"
)

func (c *CLI) stop() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("container ID required for stop")
	}
	return nil
}
