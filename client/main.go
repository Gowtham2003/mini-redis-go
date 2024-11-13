package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func main() {
	port := 1234
	if len(os.Args) == 2 {
		parsedPort, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			fmt.Println("Invalid Port Specified")
		} else {
			port = int(parsedPort)
		}
	}
	conn, err := net.Dial("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Can't connect to server:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	cReader := bufio.NewReader(conn)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Println("Received signal:", sig)
			conn.Write([]byte("QUIT\r\n"))
			os.Exit(0)
		}
	}()
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		conn.Write([]byte(strings.TrimSpace(text) + "\r\n"))

		resp, err := cReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}

		fmt.Println("Received response:", strings.TrimSpace(resp))

		if strings.TrimSpace(resp) == "ERROR" {
			fmt.Println("Key not found!")
			continue
		}

		// Check if response is a length value, it should be an integer
		v, err := strconv.ParseInt(strings.TrimSpace(resp), 10, 64)
		if err != nil {
			// If not an integer, it could be a message (e.g., OK or PONG)
			if strings.TrimSpace(resp) == "PONG" || strings.TrimSpace(resp) == "OK" {
				// Handle special cases
				continue
			}
			fmt.Println("Unexpected response:", resp)
			continue
		}

		// If the response is an integer, it represents the length of the data
		buf := make([]byte, v)
		_, err = cReader.Read(buf)
		if err != nil {
			fmt.Println("Error reading data from connection:", err)
			return
		}
		fmt.Println("Received data:", string(buf))
	}
}
