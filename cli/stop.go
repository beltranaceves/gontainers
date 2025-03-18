package cli

import (
	"fmt"
	"os"
)

func stop() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("container ID required for stop")
	}
	return nil
}
