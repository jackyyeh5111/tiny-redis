package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	port := flag.String("p", "6359", "Port of the Redis server")
	host := flag.String("h", "127.0.0.1", "Host of the Redis server")
	flag.Parse()

	// 1. Connect to our Go Redis server
	address := fmt.Sprintf("%s:%s", *host, *port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("%s> ", address)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "exit" || text == "quit" {
			break
		}

		// 2. Convert raw input (e.g., "SET key val") to RESP Array
		respCommand := convertToRESP(text)

		// 3. Send to server
		conn.Write([]byte(respCommand))

		// 4. Read the first byte of response to know how to print it
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading response:", err)
			break
		}

		// 5. Basic Parser for the display
		printResponse(response, reader)

		fmt.Print("127.0.0.1:6359> ")
	}
}

// Helper to wrap "SET key value" into "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
func convertToRESP(input string) string {
	parts := strings.Fields(input)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(parts)))
	for _, p := range parts {
		sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(p), p))
	}
	return sb.String()
}

func printResponse(res string, reader *bufio.Reader) {
	prefix := res[0]
	payload := strings.TrimSpace(res[1:])

	switch prefix {
	case '+': // Simple String
		fmt.Println(payload)
	case '-': // Error
		fmt.Println("(error)", payload)
	case ':': // Integer
		fmt.Println("(integer)", payload)
	case '$': // Bulk String
		// length, _ := fmt.Sscanf(payload, "%d", &length)
		if payload == "-1" {
			fmt.Println("(nil)")
		} else {
			// Read the actual data line
			data, _ := reader.ReadString('\n')
			fmt.Printf("\"%s\"\n", strings.TrimSpace(data))
		}
	default:
		fmt.Println(res)
	}
}
