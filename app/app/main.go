package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var port string = "6359"

func main() {
	// 1. Listen for incoming connections on port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Server is listening on port %s...\n", port)

	for {
		// 2. Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// 3. Handle the connection in a new goroutine (concurrency)
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("New connection from: %s\n", conn.RemoteAddr().String())

	// 4. Read data from the connection and write it back
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Printf("Received: %s\n", message)

		// Echo the string back to the client
		conn.Write([]byte("Echo: " + message + "\n"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading:", err)
	}
}
