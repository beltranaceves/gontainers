package main

import (
	"fmt"
	"log"
	"os"
	"gontainers/pkg/cli"
	"gontainers/pkg/daemon"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gontainers <daemon|command> [args...]")
		os.Exit(1)
	}

	if os.Args[1] == "daemon" {
		d := daemon.NewDaemon()
		if err := d.Start(); err != nil {
			log.Fatal(err)
		}
	} else {
		cmd := cli.NewCLI()
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	}
}
