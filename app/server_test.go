package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/google/uuid"
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

func TestParseHttpRequest(t *testing.T) {
	tests := []struct {
		name            string
		requestString   string
		expectedMethod  HttpMethod
		expectedPath    string
		expectedHeaders map[string]string
		expectedBody    string
	}{
		{
			name:           "RootPath",
			requestString:  "GET / HTTP/1.1\r\nHost: localhost:4221\r\nUser-Agent: curl/8.7.1\r\nAccept: */*\r\n\r\n",
			expectedMethod: GET,
			expectedPath:   "/",
			expectedHeaders: map[string]string{
				"Host":       "localhost:4221",
				"User-Agent": "curl/8.7.1",
				"Accept":     "*/*",
			},
			expectedBody: "",
		},
		{
			name:           "WithBody",
			requestString:  "POST /submit HTTP/1.1\r\nHost: localhost:4221\r\nContent-Type: application/x-www-form-urlencoded\r\nContent-Length: 13\r\n\r\nname=JohnDoe",
			expectedMethod: POST,
			expectedPath:   "/submit",
			expectedHeaders: map[string]string{
				"Host":           "localhost:4221",
				"Content-Type":   "application/x-www-form-urlencoded",
				"Content-Length": "13",
			},
			expectedBody: "name=JohnDoe",
		},
		{
			name:           "PostWithBody",
			requestString:  "POST /files/file_123 HTTP/1.1\r\nHost: localhost:4221\r\nUser-Agent: curl/8.7.1\r\nAccept: */*\r\nContent-Type: application/octet-stream\r\nContent-Length: 5\r\n\r\n12345",
			expectedMethod: POST,
			expectedPath:   "/files/file_123",
			expectedHeaders: map[string]string{
				"Host":           "localhost:4221",
				"User-Agent":     "curl/8.7.1",
				"Accept":         "*/*",
				"Content-Type":   "application/octet-stream",
				"Content-Length": "5",
			},
			expectedBody: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := parseHttpRequest(tt.requestString)
			if req.Method != tt.expectedMethod {
				t.Errorf("Expected Method %s, but got %s", tt.expectedMethod, req.Method)
			}
			if req.Path != tt.expectedPath {
				t.Errorf("Expected Path %s, but got %s", tt.expectedPath, req.Path)
			}
			for key, expectedValue := range tt.expectedHeaders {
				if req.Headers[key] != expectedValue {
					t.Errorf("Expected Header %s to be %s, but got %s", key, expectedValue, req.Headers[key])
				}
			}
			if req.Body != tt.expectedBody {
				t.Errorf("Expected Body %s, but got %s", tt.expectedBody, req.Body)
			}
		})
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
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
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
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
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
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
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

func TestGenerateHttpResponse_FileCreate(t *testing.T) {
	u := uuid.New()
	param := u.String()
	directory = os.TempDir()

	request := HttpRequest{
		Method:  POST,
		Path:    "/files/" + param,
		Headers: map[string]string{},
		Body:    "12345",
	}

	expectedStatusCode := 201
	expectedStatus := "Created"
	expectedFileContent := "12345"

	response := generateHttpResponse(request)

	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	// Check if the file was created
	filePath := filepath.Join(directory, param)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it does not", filePath)
	}

	// Check contents of the filePath equals params
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != expectedFileContent {
		t.Errorf("Expected file content to be %s, but got %s", expectedFileContent, string(content))
	}

	// Clean up
	os.Remove(filePath)

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
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
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
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
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
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
	}
	if response.Headers["Content-Length"] != strconv.Itoa(len(uaString)) {
		t.Errorf("Expected Content-Length to be %d, but got %s",
			len(uaString), response.Headers["Content-Length"])
	}
}

func TestGenerateHttpResponse_GzipEncoding(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/echo/Hello",
		Headers: map[string]string{"Accept-Encoding": "gzip"},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := "Hello"

	response := generateHttpResponse(request)
	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	responseBody, _ := decodeGzipToString(response.Body)
	if responseBody != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
	}
	if response.Headers["Content-Encoding"] != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip, but got %s",
			response.Headers["Content-Encoding"])
	}
}

func TestGenerateHttpResponse_InvalidEncoding(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/echo/Hello",
		Headers: map[string]string{"Accept-Encoding": "invalid-encoding"},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := "Hello"

	response := generateHttpResponse(request)
	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
	}
	if response.Headers["Content-Encoding"] != "" {
		t.Errorf("Expected No Content-Encoding but got %s",
			response.Headers["Content-Encoding"])
	}
}

func TestGenerateHttpResponse_MultipleEncodings_gzip_invalid(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/echo/Hello",
		Headers: map[string]string{"Accept-Encoding": "invalid-encoding-1, gzip, invalid-encoding-2"},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := "Hello"

	response := generateHttpResponse(request)
	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	responseBody, _ := decodeGzipToString(response.Body)
	if responseBody != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
	}
	if response.Headers["Content-Encoding"] != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip, but got %s",
			response.Headers["Content-Encoding"])
	}
}

func TestGenerateHttpResponse_MultipleEncodings_only_invalid(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/echo/Hello",
		Headers: map[string]string{"Accept-Encoding": "invalid-encoding-1, invalid-encoding-2"},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := "Hello"

	response := generateHttpResponse(request)
	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	if response.GetBodyAsString() != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
	}
	if response.Headers["Content-Encoding"] != "" {
		t.Errorf("Expected No Content-Encoding but got %s",
			response.Headers["Content-Encoding"])
	}
}

func TestGenerateHttpResponse_GzipEncoding_WithBody(t *testing.T) {
	request := HttpRequest{
		Method:  GET,
		Path:    "/echo/Hello",
		Headers: map[string]string{"Accept-Encoding": "gzip"},
		Body:    "",
	}

	expectedStatusCode := 200
	expectedStatus := "OK"
	expectedBody := "Hello"

	response := generateHttpResponse(request)
	if response.StatusCode != expectedStatusCode {
		t.Errorf("Expected StatusCode %d, but got %d", expectedStatusCode, response.StatusCode)
	}
	if response.Status != expectedStatus {
		t.Errorf("Expected Status %s, but got %s", expectedStatus, response.Status)
	}
	responseBody, _ := decodeGzipToString(response.Body)
	if responseBody != expectedBody {
		t.Errorf("Expected Body %s, but got %s", expectedBody, response.GetBodyAsString())
	}
	if response.Headers["Content-Encoding"] != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip, but got %s",
			response.Headers["Content-Encoding"])
	}
}

// Decode Gzip-compressed bytes back to a string
func decodeGzipToString(compressedBytes []byte) (string, error) {
	// Create a gzip reader
	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedBytes))
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()

	// Read the decompressed content
	var buf bytes.Buffer
	_, err = io.Copy(&buf, gzipReader)
	if err != nil {
		return "", err
	}

	// Return the decompressed string
	return buf.String(), nil
}
