// minicache is a work in progress in-memory caching system
// featuring a similar text based protocol to memcached/redis
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// Start the server on port 5268
	server, err := net.Listen("tcp", ":5268")

	if err != nil {
		log.Fatal(err)
	}

	// Ensure that the server closes
	defer server.Close()

	for {
		// Wait for a connection.
		connection, err := server.Accept()

		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(client net.Conn) {
			// Create a new scanner for the client input
			scanner := bufio.NewScanner(client)

			// Handle each line (command)
			for scanner.Scan() {
				// Split the command up by spaces into a slice
				input := strings.Split(scanner.Text(), " ")

				// Get the command
				command := input[0]

				// Switch on the type of command
				switch {
				case command == "get":
					// Check if a key was passed, if so try and retrieve it
					if len(input) == 2 {
						fmt.Fprintln(client, "VALUE "+command+" 0 4")
						fmt.Fprintln(client, "test")
						fmt.Fprintln(client, "END")
					} else {
						fmt.Fprintln(client, "ERROR")
					}
				case true:
					fmt.Fprintln(client, "ERROR")
				}
			}

			// Print out errors to stderr
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "[ERROR] ", err)
			}

			// Shut down the connection.
			client.Close()
		}(connection)
	}
}
