package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection Error", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection", err)
		}
	}(conn)
	// Create a byte array that would serve as a buffer
	buf := make([]byte, 1024)

	// Read from the connection
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from the connection:", err)
	}
	message := string(buf[:n])
	//fmt.Printf("Read: %v from the server", message)
	responseString := processRequest(message)

	fmt.Printf("Fetched: %v response", responseString)
	// Respond with a 200
	//resp := "HTTP/1.1 200 OK\r\n\r\n"
	resp := responseString
	_, err = conn.Write([]byte(resp))
	if err != nil {
		fmt.Println("Error while reading from the connection", err)
	}
}

func processRequest(requestString string) string {
	const UserAgentPrefix = "User-Agent:"
	var responseString string
	lines := strings.Split(requestString, "\n")
	var httpVerb, httpPath, httpVersion, userAgent string
	for i, line := range lines {
		fmt.Printf("\n Line %d: %s", i, line)
		if i == 0 {
			inputs := strings.Split(line, " ")
			for j, input := range inputs {
				if j == 0 {
					httpVerb = strings.ToUpper(input)
					fmt.Printf("\n http_verb %s", httpVerb)
				}
				if j == 1 {
					httpPath = input
					fmt.Printf("\n http_path %s", httpPath)
				}
				if j == 2 {
					httpVersion = strings.ToUpper(input)
					fmt.Printf("\n http_version %s", httpVersion)
				}
			}
		}
		if strings.HasPrefix(line, UserAgentPrefix) {
			userAgent = ExtractKV(line, UserAgentPrefix)
		}
	}
	if httpPath == "/" {
		responseString = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(httpPath, "/echo/") {
		index := strings.Index(httpPath, "/echo/")
		reqParam := httpPath[index+len("/echo/"):]

		responseString = fmt.Sprintf(
			"HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/plain\r\n"+
				"Content-Length: %d\r\n\r\n"+
				"%s", len(reqParam), reqParam)
	} else if strings.HasPrefix(httpPath, "/user-agent") {
		responseString = fmt.Sprintf(
			"HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/plain\r\n"+
				"Content-Length: %d\r\n\r\n"+
				"%s", len(userAgent), userAgent)

	} else {
		responseString = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	fmt.Printf("RESULT --->Response = \n%v", responseString)
	return responseString
}

func ExtractKV(line string, prefix string) string {
	index := strings.Index(line, prefix)
	value := line[index+len(prefix):]
	return strings.TrimSpace(value)
}
