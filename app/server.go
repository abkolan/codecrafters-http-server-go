package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var directory string

func main() {
	// Parse arguments
	directory = parseArgs()

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

func parseArgs() string {
	dir := flag.String("directory", os.TempDir(), "Directory to serve files from")
	flag.Parse()
	return *dir
}

func closeListener(ln net.Listener) {
	if err := ln.Close(); err != nil {
		fmt.Println("Error while closing the listener")
	}
}

func handleConnection(conn net.Conn) {
	// Close the connection when the function returns
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
	request := parseHttpRequest(message)
	response := generateHttpResponse(request)

	fmt.Printf("Fetched: %v response", response)
	resp := fmt.Sprintf("HTTP/1.1 %d %s\r\n", response.StatusCode, response.Status)
	for key, value := range response.Headers {
		resp += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	resp += "\r\n" + response.GetBodyAsString()
	_, err = conn.Write([]byte(resp))
	if err != nil {
		fmt.Println("Error while reading from the connection", err)
	}
}
func generateHttpResponse(request HttpRequest) HttpResponse {
	var response HttpResponse

	if request.Path == "/" {
		response = HttpResponse{
			StatusCode: 200,
			Status:     "OK",
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       []byte(""),
		}
	} else if strings.HasPrefix(request.Path, "/files/") {
		if request.Method == GET {
			fileName := strings.TrimPrefix(request.Path, "/files/")
			filePathToServe := filepath.Join(directory, fileName)
			file, err := os.Open(filePathToServe)
			if err != nil {
				response = HttpResponse{
					StatusCode: 404,
					Status:     "Not Found",
					Headers:    map[string]string{"Content-Type": "text/plain"},
					Body:       []byte("File not found"),
				}
			} else {
				defer file.Close()
				fileContent, err := os.ReadFile(filePathToServe)
				if err != nil {
					response = HttpResponse{
						StatusCode: 500,
						Status:     "Internal Server Error",
						Headers:    map[string]string{"Content-Type": "text/plain"},
						Body:       []byte("Error reading file"),
					}
				} else {
					response = HttpResponse{
						StatusCode: 200,
						Status:     "OK",
						Headers: map[string]string{
							"Content-Type":   "application/octet-stream",
							"Content-Length": fmt.Sprintf("%d", len(fileContent))},
						Body: fileContent,
					}
				}
			}
		} else if request.Method == POST {
			fileName := strings.TrimPrefix(request.Path, "/files/")
			filePathToSave := filepath.Join(directory, fileName)
			err := os.WriteFile(filePathToSave, []byte(request.Body), 0644)
			if err != nil {
				response = HttpResponse{
					StatusCode: 500,
					Status:     "Internal Server Error",
					Headers:    map[string]string{"Content-Type": "text/plain"},
					Body:       []byte("Error writing file"),
				}
			} else {
				response = HttpResponse{
					StatusCode: 201,
					Status:     "Created",
					//Headers:    map[string]string{"Content-Type": "text/plain"},
					//Body:       "File created",
				}
			}
		}
	} else if strings.HasPrefix(request.Path, "/echo/") {
		msg := strings.TrimPrefix(request.Path, "/echo/")
		responseHeaders := make(map[string]string)
		responseHeaders["Content-Type"] = "text/plain"
		responseHeaders["Content-Length"] = fmt.Sprintf("%d", len(msg))
		encodingHeader := request.Headers["Accept-Encoding"]
		needsEncoding := false

		for _, encoding := range strings.Split(encodingHeader, ",") {
			if strings.TrimSpace(encoding) == string(GZIP) {
				responseHeaders["Content-Encoding"] = string(GZIP)
				needsEncoding = true
				break
			}
		}
		if needsEncoding {
			responseHeaders["Content-Encoding"] = string(GZIP)
			gzipResponse, err := encodeStringWithGzip(msg)
			if err != nil {
				response = HttpResponse{
					StatusCode: 500,
					Status:     "Internal Server Error",
					Headers:    map[string]string{"Content-Type": "text/plain"},
					Body:       []byte("Error encoding response"),
				}
				return response
			} else {
				responseHeaders["Content-Length"] = fmt.Sprintf("%d", len(gzipResponse))
				response = HttpResponse{
					StatusCode: 200,
					Status:     "OK",
					Headers:    responseHeaders,
					Body:       gzipResponse,
				}
				return response
			}
		}

		response = HttpResponse{
			StatusCode: 200,
			Status:     "OK",
			Headers:    responseHeaders,
			Body:       []byte(msg),
		}

	} else if request.Path == "/user-agent" {
		response = HttpResponse{
			StatusCode: 200,
			Status:     "OK",
			Headers: map[string]string{"Content-Type": "text/plain",
				"Content-Length": fmt.Sprintf("%d", len(request.Headers["User-Agent"]))},
			Body: []byte(request.Headers["User-Agent"]),
		}
	} else {
		response = HttpResponse{
			StatusCode: 404,
			Status:     "Not Found",
			Headers:    map[string]string{"Content-Type": "text/plain"},
			Body:       []byte("Path not found"),
		}
	}

	return response
}

type Encoding string

const (
	GZIP Encoding = "gzip"
	UTF8 Encoding = "utf-8"
)

// HttpRequest represents an HTTP request.
type HttpMethod string

const (
	GET  HttpMethod = "GET"
	POST HttpMethod = "POST"
)

type HttpRequest struct {
	Method  HttpMethod
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
		Method:  HttpMethod(method),
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}

// HttpResponse represents an HTTP response.
type HttpResponse struct {
	StatusCode int
	Status     string
	Headers    map[string]string
	Body       []byte
	Encoding   Encoding
}

func (r *HttpResponse) GetBodyAsString() string {
	return string(r.Body)
}

func (r *HttpResponse) SetBodyFromString(body string) {
	r.Body = []byte(body)
}

// Encode string with Gzip compression.
func encodeStringWithGzip(input string) ([]byte, error) {
	var buf bytes.Buffer

	// Create a new gzip writer
	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()

	// Write the string to the gzip writer
	_, err := gzipWriter.Write([]byte(input))
	if err != nil {
		return nil, err
	}

	// Close the writer to flush the data
	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	// Return the compressed bytes
	return buf.Bytes(), nil
}
