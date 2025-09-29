package receiver

import (
	"fmt"
	"log"
	"net"
)

func ReceiveFile(ipAddr string, port int) {
	ipAddr = fmt.Sprintf("%s:%d", ipAddr, port)
	conn, err := net.Dial("tcp", ipAddr)
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer conn.Close()

	fmt.Println("Receving file from sender at", ipAddr)

	// TODO: receive file
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Error reading: %v", err)
	}
	fmt.Println("Received file path:", string(buffer[:n]))
}
