package main

import (
	"fmt"
	"net"
	"os"
)

func closeListener(ln net.Listener) {
	if err := ln.Close(); err != nil {
		fmt.Println("Error while closing the listener")
	}
}

func main() {
	fmt.Println("Starting server..")
	ln, err := net.Listen("tcp", ":4221")
	if err != nil {
		//panic(err)
		fmt.Println("Error while staring a listener", err)
		os.Exit(1)
	}
	defer closeListener(ln)

	fmt.Println("TCP Server listening at 4221")

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Connection Error", err)
	}
	handleConnection(conn)
	os.Exit(0)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Create a byte array that would serve as a buffer
	buf := make([]byte, 512)

	// Read from the connection
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from the connection:", err)
	}
	message := string(buf[:n])
	fmt.Printf("Read: %v from the server", message)

	// Respond with a 200
	resp := "HTTP/1.1 200 OK\\r\\n\\r\\n"
	_, err = conn.Write([]byte(resp))
	if err != nil {
		fmt.Println("Error while reading from the connection", err)
	}
}
