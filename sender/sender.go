package sender

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func SendFile(filepath string, port int) {
	ips, err := localIPv4s()
	if err != nil {
		log.Fatalf("failed to enumerate local IPs: %v", err)
	}
	if len(ips) == 0 {
		log.Println("warning: no non-loopback IPv4 addresses found; clients may not be able to reach this machine on the LAN")
	} else {
		fmt.Println("Connect to one of these addresses on port", port)
		for _, ip := range ips {
			fmt.Printf("  %s - [%s]\n", ip.ip, ip.ifName)
		}
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	fmt.Println()
	fmt.Println("Waiting for connection to send file...")

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("Error accepting connection: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection accepted from", conn.RemoteAddr())
	fmt.Println("Sending file:", filepath)

	// TODO: send file
	conn.Write([]byte(filepath))
}

type ifaceAddr struct {
	ifName string
	ip     string
}

func localIPv4s() ([]ifaceAddr, error) {
	var results []ifaceAddr

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		name := iface.Name
		// Check if interface is up
		if iface.Flags&net.FlagUp != net.FlagUp {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		// Get local adresses of the interface
		for _, addr := range addresses {
			ipAddr, ipNet, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, err
			}
			// Only consider IPv4 addresses that are not loopback
			if ipAddr.To4() == nil || ipNet.IP.IsLoopback() {
				continue
			}
			// only consider local addresses
			if strings.HasPrefix(ipAddr.String(), "192.168.") {
				results = append(results, ifaceAddr{ifName: name, ip: ipAddr.String()})
			}
		}
	}
	return results, nil
}
