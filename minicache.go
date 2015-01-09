// minicache is a work in progress in-memory caching system that aims
// to be wire compatible with memcached.
package main

import (
	"net"
	"bufio"
	"os"
	"log"
	"fmt"
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
				fmt.Println("[COMMAND] " + scanner.Text())
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
