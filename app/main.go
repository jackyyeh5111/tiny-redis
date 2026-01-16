package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var port string = "6359"

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
	// 2. Read length of the array
	line, _ := reader.ReadString('\n')
	count, _ := strconv.Atoi(strings.TrimSpace(line))
	fmt.Printf("line: %s \nArray of %d elements received\n", line, count)

	var args []string
	for i := 0; i < count; i++ {
		// Expecting Bulk Strings ($)
		typeByte, _ := reader.ReadByte()
		if typeByte == '$' {
			// Read Bulk String length
			line, _ := reader.ReadString('\n')
			strLen, _ := strconv.Atoi(strings.TrimSpace(line))

			// Read the actual string data
			data := make([]byte, strLen)
			io.ReadFull(reader, data)

			// Consume the trailing \r\n
			reader.ReadString('\n')

			args = append(args, string(data))
		}

		// TODO: Handle other types within the array if needed
	}

	// 3. Simple Command Logic
	if len(args) > 0 {
		command := strings.ToUpper(args[0])
		if command == "PING" {
			conn.Write([]byte("+PONG\r\n"))
		} else {
			conn.Write([]byte("-ERR unknown command\r\n"))
		}
	}
}
