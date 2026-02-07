# VPNClient

VPNClient is a companion client application for the custom Go-based VPNServer.  
It is designed to run on Windows and establish a secure VPN tunnel to the server.

The client performs encrypted handshakes, registers with the VPN server,
and forwards network traffic through a virtual tunnel interface.

This project is intended as a learning exercise focused on low-level networking,
secure communication, and protocol implementation.

---

## Key Features

- Custom VPN client written in Go
- Windows-compatible implementation
- UDP-based encrypted tunnel
- Secure handshake and session establishment
- Virtual IP address assignment managed by the server
- Encrypted packet forwarding through the VPN tunnel

---

## Architecture Overview

VPNClient is responsible for:

- Connecting to the VPN server
- Performing an encrypted handshake
- Registering as a VPN client
- Receiving a virtual IP address
- Capturing outgoing traffic from the local system
- Encrypting and forwarding packets over UDP
- Decrypting incoming packets from the VPN server

The design emphasizes transparency and simplicity to better understand
VPN internals and secure tunneling mechanisms.

---

## Configuration

Before running the client, you must configure the VPN server address.

Open `main.go` and update the server address to match your environment:

---

## Usage

Basic example:

```bash
go run main.go
