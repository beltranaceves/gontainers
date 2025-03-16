package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type CLI struct {
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
		return c.run()
	case "ps":
		return c.ps()
	case "stop":
		return c.stop()
	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}
func (c *CLI) run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("command required for run")
	}
	return nil
}

func (c *CLI) ps() error {
	type containerInfo struct {
		id      string
		command string
		pid     string
		status  string
	}

	var containers []containerInfo

	processes, err := os.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("failed to read /proc directory: %v", err)
	}

	for _, process := range processes {
		if !process.IsDir() {
			continue
		}

		cmdlinePath := fmt.Sprintf("/proc/%s/cmdline", process.Name())
		cmdline, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		cmdlineStr := strings.Join(strings.Split(string(cmdline), "\x00"), " ")
		if len(cmdline) > 0 && strings.HasPrefix(cmdlineStr, "gontainer-") {
			// Extract container ID
			containerId := strings.TrimPrefix(strings.Split(cmdlineStr, " ")[0], "gontainer-")

			// Get the command without the container prefix
			command := strings.TrimSpace(strings.TrimPrefix(cmdlineStr, "gontainer-"+containerId))
			if command == "" {
				command = "<none>"
			}

			// Get process state
			status, _ := os.ReadFile(fmt.Sprintf("/proc/%s/status", process.Name()))
			statusLines := strings.Split(string(status), "\n")
			state := "UNKNOWN"
			for _, line := range statusLines {
				if strings.HasPrefix(line, "State:") {
					stateParts := strings.SplitN(strings.TrimPrefix(line, "State:"), " ", 2)
					if len(stateParts) > 0 {
						// Just take the first character (R for running, S for sleeping, etc.)
						state = strings.TrimSpace(stateParts[0])
						// Convert to a more user-friendly format
						switch state {
						case "R":
							state = "RUNNING"
						case "S":
							state = "SLEEPING"
						case "D":
							state = "WAITING"
						case "Z":
							state = "ZOMBIE"
						case "T":
							state = "STOPPED"
						}
					}
					break
				}
			}

			containers = append(containers, containerInfo{

				id:      containerId,
				command: command,
				pid:     process.Name(),
				status:  state,
			})
		}
	}
	// Create a tabwriter with better formatting parameters
	w := tabwriter.NewWriter(os.Stdout, 12, 8, 2, ' ', 0)

	fmt.Fprintln(w, "CONTAINER ID\tCOMMAND\tPID\tSTATUS")
	for _, container := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			container.id,
			container.command,
			container.pid,
			container.status)
	}
	w.Flush()

	return nil
}
func (c *CLI) stop() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("container ID required for stop")
	}
	return nil
}
