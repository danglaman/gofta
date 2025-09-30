package sender

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func SendFiles(filepaths []string, port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listener error: %w", err)
	}
	defer ln.Close()

	ips, err := localIPv4s()
	if err != nil {
		return fmt.Errorf("failed to enumerate local IPs: %w", err)
	}

	if len(ips) == 0 {
		log.Println("warning: no non-loopback IPv4 addresses found; receiver may not be able to reach this machine")
	} else {
		fmt.Println("Connect to a address you can reach on port", port)
		for _, ip := range ips {
			fmt.Printf("  %s - [%s]\n", ip.ip, ip.ifName)
		}
	}
	fmt.Println("\nWaiting for receiver to send file(s)...")

	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("error accepting connection: %w", err)
	}
	defer conn.Close()
	fmt.Println("Connection accepted from", conn.RemoteAddr())

	// Send all files
	// protocol: send the number of files
	fmt.Fprintf(conn, "%d\n", len(filepaths))
	// 2. Send all files
	w := bufio.NewWriter(conn)
	for i, filepath := range filepaths {
		if err := sendFile(w, filepath); err != nil {
			fmt.Printf("couldn't send file %s: %v", filepath, err)
		} else {
			fmt.Printf("	sent file %s - [%d/%d]\n", filepath, i+1, len(filepaths))
		}
	}
	return nil
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
			if ipAddr.IsPrivate() {
				results = append(results, ifaceAddr{ifName: name, ip: ipAddr.String()})
			}
		}
	}
	return results, nil
}

func sendFile(w *bufio.Writer, filepath string) error {
	// Check the file
	fi, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("error parsing file %s: %w", filepath, err)
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("error: %s is not a file", filepath)
	}

	// protocol: prepend a headers: name length, name, original file size
	name := fi.Name()
	if err := binary.Write(w, binary.BigEndian, uint64(len(name))); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}
	if _, err := w.WriteString(name); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, uint64(fi.Size())); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write the file content
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filepath, err)
	}
	defer f.Close()

	// Compress the file content with gzip
	// f -> io.Copy -> zw -> w -> conn
	zw := gzip.NewWriter(w)
	if _, err := io.Copy(zw, f); err != nil {
		zw.Close()
		return fmt.Errorf("error compressing file %s: %w", filepath, err)
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("error closing gzip writer for file %s: %w", filepath, err)
	}

	return w.Flush()
}
