package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestParseArgs_success(t *testing.T) {
	os.Args = []string{"cmd", "-directory", "/tmp/foo/"}
	directory := parseArgs()
	if directory != "/tmp/foo/" {
		t.Errorf("Expected directory to be /tmp/foo/, got %s", directory)
	}
}

func TestParse_HttpRequest_Empty(t *testing.T) {
	request := "GET / HTTP/1.1\r\nHost: localhost:4221\r\nUser-Agent: curl/8.7.1\r\nAccept: */*\r\n\r\n"
	expectedMethod := HttpMethod("GET")
	expectedPath := "/"
	expectedBody := ""

	req := parseHttpRequest(request)
	if req.Method != expectedMethod {
		t.Errorf("Expected Method %s, but got %s", expectedMethod, req.Method)
	}
	if req.Path != expectedPath {
		t.Errorf("Expected Path %s, but got %s", expectedPath, req.Path)
	}
	if req.Body != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, req.Body)
	}

}
func TestGenerateHttpResponse_RootPath(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/",
		Headers: map[string]string{},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := ""

	response := generateHttpResponse(request)

	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.Body != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.Body)
	}
}

func TestGenerateHttpResponse_FileNotFound(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/files/nonexistent.txt",
		Headers: map[string]string{},
		Body:    "",
	}

	expectedStatusCode := 404
	expectedStatus := "Not Found"
	expectedBody := "File not found"

	response := generateHttpResponse(request)

	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.Body != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.Body)
	}
}

func TestGenerateHttpResponse_FileFound(t *testing.T) {
	// Create a temporary file to test
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Hello, World!"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	request := HttpRequest{
		Method:  GET,
		Path:    "/files/" + filepath.Base(tmpFile.Name()),
		Headers: map[string]string{},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := content

	// Set the directory to the temp file's directory
	directory = filepath.Dir(tmpFile.Name())

	response := generateHttpResponse(request)

	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.Body != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.Body)
	}
	if response.Headers["Content-Type"] != "application/octet-stream" {
		t.Errorf("Expected Content-Type to be application/octet-stream, but got %s",
			response.Headers["Content-Type"])
	}
	if response.Headers["Content-Length"] != strconv.Itoa(len(content)) {
		t.Errorf("Expected Content-Length to be %d, but got %s",
			len(content), response.Headers["Content-Length"])
	}
}

func TestGenerateHttpResponse_PathNotFound(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/unknown",
		Headers: map[string]string{},
		Body:    "",
	}

	expectedStatusCode := 404
	expectedStatus := "Not Found"
	expectedBody := "Path not found"

	response := generateHttpResponse(request)

	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.Body != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.Body)
	}
}

func TestGenerateHttpResponse_EchoPath(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/echo/Hello",
		Headers: map[string]string{},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := "Hello"
	expectedContentType := "text/plain"
	expectedContentLength := strconv.Itoa(len(expectedBody))

	response := generateHttpResponse(request)

	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.Body != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.Body)
	}
	if response.Headers["Content-Type"] != expectedContentType {
		t.Errorf("Expected Content-Type to be application/octet-stream, but got %s",
			response.Headers["Content-Type"])
	}
	if response.Headers["Content-Length"] != expectedContentLength {
		t.Errorf("Expected Content-Length to be %s, but got %s",
			expectedContentLength, response.Headers["Content-Length"])
	}
}

func TestGenerateHttpResponse_UserAgentHeader(t *testing.T) {
	uaString := "foobar/1.2.3"

	request := HttpRequest{
		Method:  GET,
		Path:    "/user-agent",
		Headers: map[string]string{"User-Agent": "foobar/1.2.3"},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := uaString

	response := generateHttpResponse(request)

	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.Body != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.Body)
	}
	if response.Headers["Content-Length"] != strconv.Itoa(len(uaString)) {
		t.Errorf("Expected Content-Length to be %d, but got %s",
			len(uaString), response.Headers["Content-Length"])
	}
}
