package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func (c *CLI) ps() error {
	onlineContainers := getOnlineGontainersInfo()
	offlineContainers := getOfflineGontainersInfo()
	containers := append(onlineContainers, offlineContainers...)

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

func getOfflineGontainersInfo() []containerInfo {
	// Implementation to get offline containers
	// TODO: implement getOfflineGontainersInfo
	return []containerInfo{}
}

func getOnlineGontainersInfo() []containerInfo {

	var containers []containerInfo

	processes, err := os.ReadDir("/proc")
	if err != nil {
		return containers
	}

	for _, process := range processes {
		if !process.IsDir() {
			continue
		}

		// Skip non-numeric directories (PIDs are numeric)
		if _, err := fmt.Sscanf(process.Name(), "%d", new(int)); err != nil {
			continue
		}

		cmdlinePath := fmt.Sprintf("/proc/%s/cmdline", process.Name())
		cmdlineBytes, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		// Split by null bytes to get individual arguments
		args := strings.Split(string(cmdlineBytes), "\x00")
		if len(args) == 0 || !strings.HasPrefix(args[0], "gontainer-") {
			continue
		}

		// Extract container ID from the process name
		containerId := strings.TrimPrefix(args[0], "gontainer-")

		// The actual command is the second argument (index 1) and onwards
		var command string
		if len(args) > 1 {
			// Join all remaining arguments to form the command
			command = strings.Join(args[1:len(args)-1], " ") // -1 to remove the empty string at the end
		} else {
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
	return containers
}
