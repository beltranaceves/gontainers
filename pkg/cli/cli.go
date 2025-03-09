package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

type CLI struct {
	socketPath string
}

type Command struct {
	Type    string   `json:"type"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	ID      string   `json:"id,omitempty"`
	Attach  bool     `json:"attach,omitempty"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewCLI() *CLI {
	return &CLI{
		socketPath: "/var/run/gontainers.sock",
	}
}

func (c *CLI) Execute() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("command required")
	}

	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %v", err)
	}
	defer conn.Close()

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			return fmt.Errorf("command required for run")
		}
		return c.sendCommand(conn, Command{
			Type:    "run",
			Command: os.Args[2],
			Args:    os.Args[3:],
		})

	case "list":

		return c.sendCommand(conn, Command{Type: "list"})

	case "stop":
		if len(os.Args) < 3 {
			return fmt.Errorf("container ID required for stop")
		}
		return c.sendCommand(conn, Command{
			Type: "stop",
			ID:   os.Args[2],
		})

	default:
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func (c *CLI) sendCommand(conn net.Conn, cmd Command) error {
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return fmt.Errorf("failed to send command: %v", err)
	}

	if cmd.Attach {
		// Direct connection of stdin/stdout for interactive sessions
		go io.Copy(conn, os.Stdin)
		io.Copy(os.Stdout, conn)
		return nil
	}

	// Non-interactive response handling
	decoder := json.NewDecoder(conn)
	var response Response
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("command failed: %s", response.Error)
	}
	if response.Data != "" {
		fmt.Println(response.Data)
	}

	return nil
}
