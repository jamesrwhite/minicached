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

	// Our datastore..
	data := make(map[string]string)

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
			// Ensure the client closes once we're done
			defer client.Close()

			// Define the types of data we can recieve
			var command, key, value string //, flags, length, ttl string
			var input []string
			var expecting bool

			// Create a new scanner for the client input
			scanner := bufio.NewScanner(client)

			// Handle each line (command)
			for scanner.Scan() {
				// If we aren't expecting a value then split the
				// command up by spaces into a slice
				if expecting == false {
					input = strings.Split(scanner.Text(), " ")

					// Get the command
					command = input[0]
				} else {
					// Get the value, TODO: we should be using the length here
					value = scanner.Text()
				}

				// Switch on the type of command
				switch {
				// Are we expecting a value from a set command?
				case expecting == true:
					// Store the value
					data[key] = value

					// Update expecting as we're no longer expecting a value
					expecting = false

					// Inform the client we have stored the value
					// TODO: error handling here
					fmt.Fprintln(client, "STORED")
				// get key1 [key2 .... keyn]
				case command == "get":
					// Check if a key was passed, if so try and retrieve it
					if len(input) == 2 {
						// Get the key name
						key = input[1]

						// Look up the key in our datastore
						value = data[key]

						// Did it exist?
						if value != "" {
							fmt.Fprintln(client, fmt.Sprintf("VALUE "+key+" 0 %d", len(value)))
							fmt.Fprintln(client, value)
						}

						fmt.Fprintln(client, "END")
					} else {
						fmt.Fprintln(client, "ERROR")
					}
				// set key [flags] [exptime] length [casunique] [noreply]
				case command == "set":
					// Check the right number of arguments are passed
					// TODO: currently casunique and noreply are not supported
					if len(input) == 5 {
						// Get the key name
						key = input[1]

						// Get any flags, TODO: currently ignored
						// flags = input[2]

						// Get the key TTL, TODO: currently ignored
						// ttl = input[3]

						// Get the value length, TODO: currently ignored
						// length = input[4]

						// Set that we are expecting a value
						expecting = true

						// Print out a newline while we wait for the client value
						fmt.Fprint(client, "\r\n")
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
		}(connection)
	}
}
