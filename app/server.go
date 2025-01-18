package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func closeListener(ln net.Listener) {
	if err := ln.Close(); err != nil {
		fmt.Println("Error while closing the listener")
	}
}

var directory string

func main() {
	// Define the directory flag
	dir := flag.String("directory", "/tmp/", "Directory to serve files from")
	flag.Parse()
	directory = *dir

	fmt.Printf("Starting server.. Serving files from directory: %s\n", directory)
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
	} else if strings.HasPrefix(httpPath, "/files/") {
		index := strings.Index(httpPath, "/files/")
		fileName := httpPath[index+len("/files/"):]
		filePathToServe := filepath.Join(directory, fileName)
		file, err := os.Open(filePathToServe)
		if err != nil {
			responseString = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			fileInfo, err := file.Stat()
			if err != nil {
				responseString = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
			} else {
				fileSize := fileInfo.Size()
				buf := make([]byte, fileSize)
				_, err := file.Read(buf)
				if err != nil {
					responseString = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				} else {
					responseString = fmt.Sprintf(
						"HTTP/1.1 200 OK\r\n"+
							"Content-Type: application/octet-stream\r\n"+
							"Content-Length: %d\r\n\r\n"+"%s", fileSize, buf)
				}
			}
		}

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

// HttpRequest represents an HTTP request
type HttpRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

func parseHttpRequest(requestString string) HttpRequest {
	lines := strings.Split(requestString, "\r\n")
	requestLine := strings.Fields(lines[0])
	method := requestLine[0]
	path := requestLine[1]

	headers := make(map[string]string)
	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		headerParts := strings.SplitN(line, ": ", 2)
		headers[headerParts[0]] = headerParts[1]
	}

	body := strings.Join(lines[len(headers)+2:], "\r\n")

	return HttpRequest{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}
