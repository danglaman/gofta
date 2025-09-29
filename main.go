package main

import (
	"flag"

	"github.com/danglaman/gofta/receiver"
	"github.com/danglaman/gofta/sender"
)

func main() {
	var recieveFlag = flag.String("r", "", "To receive a file, provide sender's IP address")
	var sendFlag = flag.String("s", "", "To send a file, provide the file path")
	var portFlag = flag.Int("p", 5005, "Port to use for sending/receiving file. Default is 5005")
	flag.Parse()

	if *recieveFlag != "" {
		receiver.ReceiveFile(*recieveFlag, *portFlag)
		return
	} else if *sendFlag != "" {
		sender.SendFile(*sendFlag, *portFlag)
		return
	} else {
		flag.Usage()
	}
}
