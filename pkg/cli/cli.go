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

		if len(cmdline) > 0 && strings.HasPrefix(string(cmdline), "gontainer-") {
			// Clean command string
			cmdlineStr := strings.Join(strings.Split(string(cmdline), "\x00"), " ")

			// Get process state
			status, _ := os.ReadFile(fmt.Sprintf("/proc/%s/status", process.Name()))
			statusLines := strings.Split(string(status), "\n")
			state := "UNKNOWN"
			for _, line := range statusLines {
				if strings.HasPrefix(line, "State:") {
					state = strings.TrimSpace(strings.TrimPrefix(line, "State:"))
					break
				}
			}

			containers = append(containers, containerInfo{
				id:      strings.TrimPrefix(cmdlineStr, "gontainer-"),
				command: cmdlineStr,
				pid:     process.Name(),
				status:  state,
			})
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

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
