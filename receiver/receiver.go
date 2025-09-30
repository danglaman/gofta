package receiver

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ReceiveFiles(senderIP string, port int, destDir string) error {
	ipAddr := net.JoinHostPort(senderIP, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", ipAddr)
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}
	defer conn.Close()

	absPath, err := ensureDir(destDir)
	if err != nil {
		return fmt.Errorf("error creating the destination folder: %w", err)
	}

	// receive the files
	r := bufio.NewReader(conn)
	// protocol: first the number of files is sent as a
	nb, err := r.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("error reading file count: %w", err)
	}

	nFiles, err := strconv.Atoi(strings.TrimSuffix(string(nb), "\n"))
	if err != nil {
		return fmt.Errorf("protocol error: %w", err)
	}

	fmt.Printf("Receving %d file(s) from sender at %s...\n", nFiles, ipAddr)
	for i := range nFiles {
		if fName, err := receiveFile(r, absPath); err != nil {
			return fmt.Errorf("error receiving file %d: %w", i+1, err)
		} else {
			fmt.Printf("	received %s - [%d/%d]\n", fName, i+1, nFiles)
		}
	}
	return nil
}

func ensureDir(destDir string) (string, error) {
	abs, err := filepath.Abs(destDir)
	if err != nil {
		return "", fmt.Errorf("ensureDir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil { // rwxr-xr-x
		return "", fmt.Errorf("ensureDir: %w", err)
	}
	return abs, nil
}

func receiveFile(r *bufio.Reader, dstDir string) (string, error) {
	// protocol: receive name lenth [uint64]
	var nameLen uint64
	if err := binary.Read(r, binary.BigEndian, &nameLen); err != nil {
		return "", fmt.Errorf("error reading file name length: %w", err)
	}
	// protocol: receive name [string]
	fnBuf := make([]byte, nameLen)
	if _, err := io.ReadFull(r, fnBuf); err != nil {
		return "", fmt.Errorf("error reading file name: %w", err)
	}
	fn := string(fnBuf)
	// protocol: receive original file size [uint64]
	var fSize uint64
	if err := binary.Read(r, binary.BigEndian, &fSize); err != nil {
		return "", fmt.Errorf("error reading file size: %w", err)
	}
	// protocol: receive compressed file
	zr, err := gzip.NewReader(r)
	if err != nil {
		return "", fmt.Errorf("error decompressing: %w", err)
	}
	zr.Multistream(false) // stop after first EOF
	defer zr.Close()

	dstFilepath := filepath.Join(dstDir, fn)
	f, err := os.OpenFile(dstFilepath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644) //rw-r--r--
	if err != nil {
		return "", fmt.Errorf("error creating file %s: %w", dstFilepath, err)
	}

	written, copyErr := io.Copy(f, zr)
	f.Close()
	if copyErr != nil {
		_ = os.Remove(dstFilepath)
		return "", fmt.Errorf("error copying decompressed file: %w", copyErr)
	}

	// check file
	if written != int64(fSize) {
		// remove the file and throw an error
		os.Remove(dstFilepath)
		return "", fmt.Errorf("size of received file does not match size of sent file: got %d, expected %d", written, fSize)
	}

	return fn, nil
}
