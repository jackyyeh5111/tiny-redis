package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var port string = "6359"

// Define Global Storage and Mutex for Thread-Safety
var (
	store = make(map[string]string)
	mu    sync.RWMutex
)

func main() {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Server is listening on port %s...\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleRequest(conn)
	}
}

/*
RESP (REdis Serialization Protocol)

For example, the command PING is sent as: *1\r\n$4\r\nPING\r\n

Type, Prefix, Example
----------------------------------------
Simple String, +, +OK\r\n
Error, -, -Error message\r\n
Integer, :, :1000\r\n
Bulk String, $, $4\r\nPING\r\n
Array, *, *1\r\n[followed by elements]
*/
func handleRequest(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// 1. Read the first byte to determine the type (e.g., '*' for Array)
		typeByte, err := reader.ReadByte()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading type byte:", err)
			}
			return
		}

		if typeByte == '*' {
			handleArray(reader, conn)
		}

		// TODO: Handle other RESP types if needed
	}
}

func handleArray(reader *bufio.Reader, conn net.Conn) {
	line, _ := reader.ReadString('\n')
	count, _ := strconv.Atoi(strings.TrimSpace(line))

	var args []string
	for i := 0; i < count; i++ {
		typeByte, _ := reader.ReadByte()
		if typeByte == '$' {
			line, _ := reader.ReadString('\n')
			strLen, _ := strconv.Atoi(strings.TrimSpace(line))

			data := make([]byte, strLen)
			io.ReadFull(reader, data)
			reader.ReadString('\n') // consume CRLF

			args = append(args, string(data))
		}
	}

	if len(args) > 0 {
		command := strings.ToUpper(args[0])

		switch command {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))

		case "SET":
			if len(args) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				return
			}
			key, value := args[1], args[2]

			mu.Lock() // Lock for writing
			store[key] = value
			mu.Unlock()

			conn.Write([]byte("+OK\r\n"))

		case "GET":
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				return
			}
			key := args[1]

			mu.RLock() // Lock for reading (multiple readers allowed)
			value, exists := store[key]
			mu.RUnlock()

			if !exists {
				// Nil Bulk String response in RESP
				conn.Write([]byte("$-1\r\n"))
			} else {
				// Bulk String response: $[len]\r\n[value]\r\n
				response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
				conn.Write([]byte(response))
			}

		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}
