package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("beacon search <query>")
			os.Exit(1)
		}
		query := os.Args[2]
		fmt.Printf("Searching for: %s\n", query)
	case "list":
		fmt.Println("Listing notes...")
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("beacon create <name>")
			os.Exit(1)
		}
		name := os.Args[2]
		fmt.Printf("Creating note: %s\n", name)
	case "version":
		fmt.Println("beacon v0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Beacon - Headless Obsidian Vault CLI

Usage:
  beacon search <query>      Search notes by content
  beacon list [inbox|tags]   List notes
  beacon create <name>       Create new note
  beacon version             Show version

Examples:
  beacon search "golang"
  beacon list inbox
  beacon create "My Note"
`)
}
