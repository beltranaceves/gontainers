package main

import (
	"fmt"
	"gontainers/pkg/cli"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gontainers <command> [args...]")
		os.Exit(1)
	}

	cmd := cli.NewCLI()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
