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

func closeConnection(conn net.Conn) {
	if err := conn.Close(); err != nil {
		fmt.Println("Error while closing the connection", conn)
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
	defer closeConnection(conn)
	os.Exit(0)
}
