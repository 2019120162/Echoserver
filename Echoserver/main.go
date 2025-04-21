package main

import (
	"bufio"
	"flag" // Used for parsing command-line flags
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	maxMessageSize    = 1024             // Max size of a single client message
	inactivityTimeout = 30 * time.Second // Timeout duration for client inactivity
)

func main() {
	// === Command Line Flag Parsing ===
	// Define a port flag with default value "4000"
	port := flag.String("port", "4000", "TCP port to listen on")
	flag.Parse() // Parse the flags

	// Create address string using specified port
	address := fmt.Sprintf(":%s", *port)

	// Start listening on the specified TCP address
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Server listening on %s\n", address)

	// Accept incoming connections in a loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		// Handle each connection concurrently
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()

	// Log connection established
	logConnection(clientAddr, true)
	// Log disconnection upon function exit
	defer logConnection(clientAddr, false)

	// Create a per-client log file for message history
	logFile := createClientLogFile(clientAddr)
	defer logFile.Close()
	writer := bufio.NewWriter(logFile)

	// Use a scanner to read lines from the connection
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, maxMessageSize), maxMessageSize)

	// Channel used to reset inactivity timer
	timeoutChan := make(chan struct{})
	go startInactivityTimer(conn, timeoutChan)

	for {
		// Reset read deadline for inactivity timeout
		conn.SetReadDeadline(time.Now().Add(inactivityTimeout))
		if !scanner.Scan() {
			// Handle client disconnect or error
			return
		}
		// Reset inactivity timer on new input
		timeoutChan <- struct{}{}

		// Get the input string and clean it
		input := strings.TrimSpace(scanner.Text())
		if len(input) > maxMessageSize {
			conn.Write([]byte("Message too long. Max 1024 bytes allowed.\n"))
			continue
		}

		// Log message to file
		logMessage(writer, input)

		// === Simple Personality Response System ===
		switch strings.ToLower(input) {
		case "":
			conn.Write([]byte("Say something...\n"))
		case "hello":
			conn.Write([]byte("Hi there!\n"))
		case "bye":
			conn.Write([]byte("Goodbye!\n"))
			return
		default:
			// === Custom Command Handling ===
			if strings.HasPrefix(input, "/") {
				handleCommand(conn, input)
			} else {
				// Echo back general messages
				conn.Write([]byte(fmt.Sprintf("%s\n", input)))
			}
		}
	}
}

// Logs when a client connects or disconnects
func logConnection(addr string, connected bool) {
	timestamp := time.Now().Format(time.RFC3339)
	status := "connected"
	if !connected {
		status = "disconnected"
	}
	fmt.Printf("[%s] %s %s\n", timestamp, addr, status)
}

// Creates a uniquely named log file for each client based on their address
func createClientLogFile(addr string) *os.File {
	safeAddr := strings.ReplaceAll(addr, ":", "_")
	filename := fmt.Sprintf("client_%s.log", safeAddr)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error creating log file for %s: %v\n", addr, err)
	}
	return file
}

// Appends timestamped messages to the client’s log file
func logMessage(writer *bufio.Writer, msg string) {
	timestamp := time.Now().Format(time.RFC3339)
	writer.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
	writer.Flush()
}

// Processes user-issued slash (/) commands
func handleCommand(conn net.Conn, input string) {
	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case "/time":
		// Respond with server time
		now := time.Now().Format("15:04:05")
		conn.Write([]byte(fmt.Sprintf("Server time: %s\n", now)))
	case "/quit":
		// Close connection gracefully
		conn.Write([]byte("Goodbye!\n"))
		conn.Close()
	case "/echo":
		// Repeat back user message
		if len(parts) > 1 {
			conn.Write([]byte(fmt.Sprintf("%s\n", strings.Join(parts[1:], " "))))
		} else {
			conn.Write([]byte("Usage: /echo <message>\n"))
		}
	default:
		// Unknown command handler
		conn.Write([]byte("Unknown command\n"))
	}
}

// Watches for inactivity and disconnects the client if idle too long
func startInactivityTimer(conn net.Conn, resetChan chan struct{}) {
	timer := time.NewTimer(inactivityTimeout)

	for {
		select {
		case <-resetChan:
			// Reset timer when activity is detected
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(inactivityTimeout)
		case <-timer.C:
			// Trigger timeout response and close connection
			conn.Write([]byte("Disconnected due to inactivity.\n"))
			conn.Close()
			return
		}
	}
}
