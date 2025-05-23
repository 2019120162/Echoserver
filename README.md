# How to Run the Server

    Make sure you have Go installed.

    Save the code above into a file named main.go.

    In terminal:

go run main.go --port=4000

Or to use a different port:

    go run main.go --port=5000

# Functional Highlights

Here’s how each requirement is met:
Feature: Status	Description
Concurrency:	go handleConnection(conn) spawns goroutines.
Connection Logging:	IP and timestamps logged with connection/disconnection.
Graceful Disconnect:	Uses scanner.Scan() and defer conn.Close() safely.
Trimmed Input & Echo:	strings.TrimSpace() before processing.
Client Message Logging:	Saves messages to per-client log file.
Command Line Port Flag:	--port via flag.String.
Inactivity Timeout (30s):	Goroutine with time.Timer kills connection.
Overflow Protection (1024 bytes):	Rejects message over 1024 bytes.
Server Personality Mode:	Responses for hello, bye, and empty message.
Command Protocol:	Handles /time, /quit, and /echo.

# Enhanced TCP Echo Server in Go

## How to Run

go run main.go --port=4000

Or change port:

go run main.go --port=5000

Connect with netcat:

nc localhost 4000

## Demo Video:

Watch the server in action: [YouTube Demo Link](https://youtu.be/oJ2y51QGY6s)


# Most Educationally Enriching Feature

Implementing the inactivity timeout using goroutines and channels was the most enriching. It helped me understand timer behavior, conditions, and concurrent cleanup.

# Most Research-Intensive Feature

The command-line flag parsing and integrating with custom logging per IP required the most research especially ensuring safe filenames and concurrent file writes.
