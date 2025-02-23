package daemon

import (
	"encoding/json"
	"fmt"
	"gontainers/pkg/container"
	"net"
	"os"
)

type Daemon struct {
	containers map[string]*container.Container
	socketPath string
}

type Command struct {
	Type    string   `json:"type"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	ID      string   `json:"id,omitempty"`
}

type Response struct {
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewDaemon() *Daemon {
	return &Daemon{
		containers: make(map[string]*container.Container),
		socketPath: "/var/run/gontainers.sock",
	}
}

func (d *Daemon) Start() error {
	if err := os.RemoveAll(d.socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %v", err)
	}

	listener, err := net.Listen("unix", d.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go d.handleConnection(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var cmd Command
	if err := decoder.Decode(&cmd); err != nil {
		encoder.Encode(Response{Success: false, Error: err.Error()})
		return
	}

	switch cmd.Type {
	case "run":
		cont := container.NewContainer(cmd.Command, cmd.Args)
		d.containers[cont.ID] = cont
		err := cont.Start()
		if err != nil {
			encoder.Encode(Response{Success: false, Error: err.Error()})
			return
		}
		encoder.Encode(Response{Success: true, Data: cont.ID})

	case "list":
		var containerList string
		for id, c := range d.containers {
			containerList += fmt.Sprintf("ID: %s, Command: %s\n", id, c.Command)
		}
		encoder.Encode(Response{Success: true, Data: containerList})

	case "stop":
		if cont, exists := d.containers[cmd.ID]; exists {
			err := cont.Kill()
			if err != nil {
				encoder.Encode(Response{Success: false, Error: err.Error()})
				return
			}
			delete(d.containers, cmd.ID)
			encoder.Encode(Response{Success: true})
		} else {
			encoder.Encode(Response{Success: false, Error: "container not found"})
		}
	}
}
