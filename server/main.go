package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type store struct {
	mutex sync.RWMutex
	data  map[string]string
}

func (s *store) read(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	return value, exists
}

func (s *store) write(key string, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value
}

func handleRequest(conn net.Conn, st *store) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	ln, _, err := reader.ReadLine()
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	req := strings.Split(strings.TrimSpace(string(ln)), " ")

	if len(req) == 2 && req[0] == "GET" {
		fmt.Println("Received a GET Request")
		st.mutex.RLock()
		defer st.mutex.RUnlock()

		data, exists := st.read(req[1])
		if !exists {
			conn.Write([]byte("ERROR\r\n"))
		} else {
			conn.Write([]byte(fmt.Sprintf("%d\r\n", len(data)))) // Send length first
			conn.Write([]byte(data))                             // Send data
			conn.Write([]byte("\r\n"))                           // Ensure newline after data
		}

	} else if len(req) == 2 && req[0] == "DEL" {
		fmt.Println("Received a DEL Request")
		st.mutex.Lock()
		defer st.mutex.Unlock()

		delete(st.data, req[1])
		conn.Write([]byte("OK\r\n"))

	} else if len(req) == 3 && req[0] == "SET" {
		fmt.Println("Received a SET Request")
		st.mutex.Lock()
		defer st.mutex.Unlock()

		st.write(req[1], req[2])
		conn.Write([]byte("OK\r\n"))

	} else if len(req) == 1 && req[0] == "PING" {
		fmt.Println("Received a PING Request")
		conn.Write([]byte("PONG\r\n"))
	}
}

func main() {
	st := &store{
		data: make(map[string]string),
	}
	port := 1234
	if len(os.Args) == 2 {
		parsedPort, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			fmt.Println("Invalid Port Specified")
		} else {
			port = int(parsedPort)
		}
	}
	fmt.Println("Starting Mini Redis Server on localhost:" + fmt.Sprintf("%d", port))

	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Couldn't start the Mini Redis Server:", err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleRequest(conn, st)
	}
}
