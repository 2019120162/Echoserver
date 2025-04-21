package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	maxMessageSize    = 1024
	inactivityTimeout = 30 * time.Second
)

func main() {
	port := flag.String("port", "4000", "TCP port to listen on")
	flag.Parse()

	address := fmt.Sprintf(":%s", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Server listening on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	logConnection(clientAddr, true)
	defer logConnection(clientAddr, false)

	logFile := createClientLogFile(clientAddr)
	defer logFile.Close()
	writer := bufio.NewWriter(logFile)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, maxMessageSize), maxMessageSize)

	timeoutChan := make(chan struct{})
	go startInactivityTimer(conn, timeoutChan)

	for {
		conn.SetReadDeadline(time.Now().Add(inactivityTimeout))
		if !scanner.Scan() {
			// handle client disconnect
			return
		}
		timeoutChan <- struct{}{}

		input := strings.TrimSpace(scanner.Text())
		if len(input) > maxMessageSize {
			conn.Write([]byte("Message too long. Max 1024 bytes allowed.\n"))
			continue
		}

		logMessage(writer, input)

		// Personality mode
		switch strings.ToLower(input) {
		case "":
			conn.Write([]byte("Say something...\n"))
		case "hello":
			conn.Write([]byte("Hi there!\n"))
		case "bye":
			conn.Write([]byte("Goodbye!\n"))
			return
		default:
			// Command protocol
			if strings.HasPrefix(input, "/") {
				handleCommand(conn, input)
			} else {
				conn.Write([]byte(fmt.Sprintf("%s\n", input)))
			}
		}
	}
}

func logConnection(addr string, connected bool) {
	timestamp := time.Now().Format(time.RFC3339)
	status := "connected"
	if !connected {
		status = "disconnected"
	}
	fmt.Printf("[%s] %s %s\n", timestamp, addr, status)
}

func createClientLogFile(addr string) *os.File {
	safeAddr := strings.ReplaceAll(addr, ":", "_")
	filename := fmt.Sprintf("client_%s.log", safeAddr)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error creating log file for %s: %v\n", addr, err)
	}
	return file
}

func logMessage(writer *bufio.Writer, msg string) {
	timestamp := time.Now().Format(time.RFC3339)
	writer.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
	writer.Flush()
}

func handleCommand(conn net.Conn, input string) {
	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case "/time":
		now := time.Now().Format("15:04:05")
		conn.Write([]byte(fmt.Sprintf("Server time: %s\n", now)))
	case "/quit":
		conn.Write([]byte("Goodbye!\n"))
		conn.Close()
	case "/echo":
		if len(parts) > 1 {
			conn.Write([]byte(fmt.Sprintf("%s\n", strings.Join(parts[1:], " "))))
		} else {
			conn.Write([]byte("Usage: /echo <message>\n"))
		}
	default:
		conn.Write([]byte("Unknown command\n"))
	}
}

func startInactivityTimer(conn net.Conn, resetChan chan struct{}) {
	timer := time.NewTimer(inactivityTimeout)

	for {
		select {
		case <-resetChan:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(inactivityTimeout)
		case <-timer.C:
			conn.Write([]byte("Disconnected due to inactivity.\n"))
			conn.Close()
			return
		}
	}
}
