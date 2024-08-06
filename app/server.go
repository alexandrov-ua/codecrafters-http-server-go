package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// urls
var urls = make(map[string]int)

func main() {
	urls["/"] = 1
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()

	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	if err := AcceptConnection(conn); err != nil {
		fmt.Println(err)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
	}
}

func AcceptConnection(conn net.Conn) error {

	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}
	reader := bufio.NewReader(bytes.NewReader(buff[:n]))
	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}
	_, path, err := ParseStartLine(line)
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}

	if _, ok := urls[path]; ok {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	return nil
}

func ParseStartLine(l []byte) (verb string, path string, err error) {
	arr := strings.Split(string(l), " ")
	if len(arr) != 3 || (arr[2] != "HTTP/1.1" && arr[2] != "HTTP/1.0") {
		return "", "", errors.New("not a http")
	}
	return arr[0], arr[1], nil
}
