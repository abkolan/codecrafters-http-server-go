# Build your own HTTP Server from scratch challenge
[![progress-banner](https://backend.codecrafters.io/progress/http-server/a74b3908-d053-4cb0-b8e6-762e0bfb3855)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

[![Go Report Card](https://goreportcard.com/badge/github.com/abkolan/codecrafters-http-server-go)](https://goreportcard.com/report/github.com/abkolan/codecrafters-http-server-go)  

This is part of [codecrafters.io's](https://codecrafters.io) 
["Build Your Own HTTP server" Challenge](https://app.codecrafters.io/courses/http-server/overview)  
Had fun building this one.

## How to run
1. Ensure you have `go (1.19)` installed locally
2. Run `./your_server.sh` to run the HttpServer which is implemented in
   `app/server.go`.
3. Or run `make build` or `make test` 

## Features
### Respond with a 200 on root path
Request
```bash
$ curl -v http://localhost:4221
```
Response
```http
HTTP/1.1 200 OK
Content-Type: text/plain


```
### Echo Server
Request 
```
$ curl -v http://localhost:4221/echo/abc
```
Response 
```http
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 3

abc
```
### Reads Content Headers and returns `User_Agent` as a response
Request 
```bash
$ curl -v --header "User-Agent: foobar/1.2.3" http://localhost:4221/user-agent
```
Response
```http
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 12

foobar/1.2.3
```
### Supports concurrent connections
Requests
```bash
$ (sleep 3 && printf "GET / HTTP/1.1\r\n\r\n") | nc localhost 4221 &
$ (sleep 3 && printf "GET / HTTP/1.1\r\n\r\n") | nc localhost 4221 &
$ (sleep 3 && printf "GET / HTTP/1.1\r\n\r\n") | nc localhost 4221 &
```
Responses
```http
HTTP/1.1 200 OK
Content-Type: text/plain

HTTP/1.1 200 OK
Content-Type: text/plain

HTTP/1.1 200 OK
Content-Type: text/plain

```

### Support File Downloads
Request 1
```bash
$ echo -n 'Hello, World!' > $TMPDIR/foo
$ curl -i http://localhost:4221/files/foo
```
Response 1
```http
HTTP/1.1 200 OK
Content-Type: application/octet-stream
Content-Length: 13

Hello, World!

```
Request 2

```bash
$ curl -i http://localhost:4221/files/non_existant_file
```
Response 2
```http
HTTP/1.1 404 Not Found
Content-Type: text/plain

File not found
```
### Supports File Uploads
Request 1
```bash
curl -v --data "12345" -H "Content-Type: application/octet-stream" http://localhost:4221/files/file_123
```
Response 1
```http
HTTP/1.1 201 Created

```
Request 2
```bash
curl -i http://localhost:4221/files/file_123 
```
Response 2 
```http
HTTP/1.1 200 OK
Content-Type: application/octet-stream
Content-Length: 5

12345
```
### Supports gzip compression 
Request 
```bash
curl -v -H "Accept-Encoding: gzip" http://localhost:4221/echo/abc
```
Response (formatted)
```http
HTTP/1.1 200 OK
Content-Length: 27
Content-Encoding: gzip
Content-Type: text/plain
 
{ [27 bytes data]
100    27  100    27    0     0   4373      0 --:--:-- --:--:-- --:--:--  4500
* Connection #0 to host localhost left intact
00000000  1f 8b 08 00 00 00 00 00  00 ff 4a 4c 4a 06 04 00  |..........JLJ...|
00000010  00 ff ff c2 41 24 35 03  00 00 00                 |....A$5....|
0000001b
```
