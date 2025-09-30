# gofta – Go File Transfer App

A simple TCP-based file transfer utility written in Go.  
It supports sending multiple files and receiving them on another machine in the same network.

## Features

- **Send and receive** multiple files over TCP sockets    
- **File compression** using `compress/gzip` for less network usage
- Implemented entirely with the **Go standard library**
- **Cross-platform**

## Security Warning

All transfers are sent in **plain TCP without encryption or authentication**.  
Do **not** use this tool to send sensitive, confidential, or personal files over untrusted networks.  
It is designed for simple transfers within a trusted local network only.


## Installation

Clone and build:

```bash
git clone https://github.com/danglaman/gofta.git
cd gofta
go build
```
This will create the binary `gofta` in your working directory

## Usage
### Sender
Run the sender to listen for a connection and wait until the receiver connects:
```bash
./gofta send [--port <port>] <file1> [<file2> ...]
```
Example:
```bash
./gofta send file1.txt file2.pdf
```

Output:
```bash
Connect to a address you can reach on port 5005
  192.168.56.1 - [Ethernet 2]
  192.168.178.23 - [WLAN]

Waiting for receiver to send file(s)...
Connection accepted from 192.168.178.151:57942
        sent file file1.txt - [1/2]
        sent file file2.pdf - [2/2]
```

### Receiver
Run the receiver to connect to the sender’s IP and specify a destination path:
```bash
./gofta receive [--port <port>] <ip-address> <path>
```

Example:
```bash
./gofta receive 192.168.178.151 .
```
Output:
```bash
Receving 2 file(s) from sender at 192.168.178.151:5005...
        received file1.txt - [1/2]
        received file2.pdf - [2/2]
```
### Defaults
- Default port: `5005`
- Sender prints the available local IPs and network interfaces.
- Files are stored into the given `<path>` on the receiver side.
- Files with the same name **will be overwritten**.
- The destination directory will be created if it doesn’t exist.

## Protocol
The sender and receiver communicate in a simple custom protocol:

1. Sender writes the number of files as a line.
2. For each file:
    - 8-byte (uint64, big endian) filename length
    - filename (bytes)
    - 8-byte (uint64, big endian) original file size
    - gzip-compressed file content

Receiver reads the headers, decompresses the file content, writes to disk, and verifies the uncompressed size.