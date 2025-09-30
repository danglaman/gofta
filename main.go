package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/danglaman/gofta/receiver"
	"github.com/danglaman/gofta/sender"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "send":
		subargs := os.Args[2:]
		fs := flag.NewFlagSet("send", flag.ExitOnError)
		port := fs.Int("port", 5005, "Port to listen on")

		fs.Parse(subargs)
		files := fs.Args()
		if len(files) == 0 {
			fmt.Println("error: missing file path(s) to send")
			printUsage()
			os.Exit(2)
		}
		if err := sender.SendFiles(files, *port); err != nil {
			fmt.Println("error sending files:", err)
			os.Exit(1)
		}
	case "receive":
		subargs := os.Args[2:]
		fs := flag.NewFlagSet("receive", flag.ExitOnError)
		port := fs.Int("port", 5005, "TCP port to connect to on sender")

		fs.Parse(subargs)
		args := fs.Args()
		if len(args) < 2 {
			printUsage()
			os.Exit(2)
		}
		ip := args[0]
		destDir := args[1]
		if err := receiver.ReceiveFiles(ip, *port, destDir); err != nil {
			fmt.Println("error receiving files:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("error: unknown action", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	const usage = `Usage:
	gofta send [--port <port>] <file1> [<file2> ...]
		Sender listens and waits for receiver to connect.

	gofta receive [--port <port>] <ip-address> <path>
		Connect to sender and receive files. Files are saved into <path>.

	Default port is 5005
	`
	fmt.Println(usage)
}
