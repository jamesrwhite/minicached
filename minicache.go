// minicache is a work in progress in-memory caching system
// featuring a similar text based protocol to memcached
package main

// TODO:
// - TTL handling
// - casunique?
// - noreply?

import (
	"bufio"
	"fmt"
	"log"
	"net"
	// "os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Id      string
	State   uint8
	Input   []string
	Command string
	Record  *Record
}

type Record struct {
	Key    string
	Value  string
	Flags  int64
	Ttl    int64
	Length int64
}
	
const STATE_DEFAULT uint8 = 1
const STATE_COMMAND_GET uint8 = 2
const STATE_COMMAND_SET uint8 = 3
const STATE_EXPECTING_VALUE uint8 = 4

var ticker = time.NewTicker(time.Second * 1)

func main() {
	// Start the server on port 5268
	server, err := net.Listen("tcp", ":5268")

	if err != nil {
		log.Fatal(err)
	}

	// Ensure that the server closes
	defer server.Close()

	// Our datastore..
	datastore := make(map[string]*Record)

	// A list of active clients
	clients := make(map[string]*Client)

	go func() {
		for range ticker.C {
			for key, value := range datastore {
				fmt.Println(key, value)
			}
		}
	}()

	for {
		// Wait for a connection.
		connection, err := server.Accept()

		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting new connections,
		// so that multiple connections may be served concurrently.
		go func(connection net.Conn) {
			// Create the client and store it
			client := &Client{
				Id:      connection.RemoteAddr().String(),
				State:   STATE_DEFAULT,
				Input:   []string{},
				Command: "",
				Record:  &Record{},
			}
			clients[connection.RemoteAddr().String()] = client

			// Ensure the client conncetion closes once we're done
			defer connection.Close()

			// Create a new scanner for the client input
			scanner := bufio.NewScanner(connection)

			// Handle each line (command)
			for scanner.Scan() {
				// Split the client input up based on spaces
				client.Input = strings.Split(scanner.Text(), " ")

				// If we're in our default state then determine
				// what command we're running
				if client.State == STATE_DEFAULT {
					// Get the command
					client.Command = client.Input[0]

					switch client.Command {
					case "get":
						client.State = STATE_COMMAND_GET
					case "set":
						client.State = STATE_COMMAND_SET
					}
				}

				// Switch on the type of command
				switch client.State {
				// Are we expecting a value from a set command?
				case STATE_EXPECTING_VALUE:
					// If the value isn't set then set it
					if client.Record.Value == "" {
						client.Record.Value = scanner.Text()
					// Otherwise append to it
					} else {
						client.Record.Value += scanner.Text()
					}

					client.Record.Value += "\r\n"

					// Count the length of the value minus the trailing \r\n
					valueLength := int64(len(client.Record.Value)) - 2

					// If the datastore is same or greater than the expected length
					// we are done with this op
					if valueLength >= client.Record.Length {
						// If it's the same length we can try and store it
						if valueLength == client.Record.Length {
							// Store the value
							datastore[client.Record.Key] = client.Record

							// Inform the client we have stored the value
							// TODO: error handling here
							fmt.Fprintln(connection, "STORED")
						// Otherwise the client has messed up
						} else {
							// Inform the client that they messed up
							fmt.Fprintln(connection, "CLIENT_ERROR")
							fmt.Fprintln(connection, valueLength)
							fmt.Fprintln(connection, "ERROR")
						}

						// Reset the clients state
						client.Reset()
					}
				// get key1 [key2 .... keyn]
				case STATE_COMMAND_GET:
					// Check if a key was passed, if so try and retrieve it
					if len(client.Input) == 2 {
						// Get the key
						key := client.Input[1]

						// Look up the record in our datastore
						record := datastore[key]

						// Did it exist?
						if record != nil {
							fmt.Fprintln(connection, fmt.Sprintf("VALUE %s %d %d", record.Key, record.Flags, record.Length))
							fmt.Fprint(connection, record.Value)
						}

						fmt.Fprintln(connection, "END")
					} else {
						fmt.Fprintln(connection, "CLIENT_ERROR")
					}

					// Reset the clients state regardless off success/failure
					client.Reset()
				// set key [flags] [exptime] length [casunique] [noreply]
				case STATE_COMMAND_SET:
					// Check the right number of arguments are passed
					if len(client.Input) == 5 {
						// Get the key name
						client.Record.Key = client.Input[1]

						// Get any flags
						client.Record.Flags, err = strconv.ParseInt(client.Input[2], 10, 64)

						if err != nil {
							fmt.Fprintln(connection, "CLIENT_ERROR ", err)
							break
						}

						// Get the key TTL
						client.Record.Ttl, err = strconv.ParseInt(client.Input[3], 10, 64)

						if err != nil {
							fmt.Fprintln(connection, "CLIENT_ERROR ", err)
							break
						}

						// Get the value length
						client.Record.Length, err = strconv.ParseInt(client.Input[4], 10, 64)

						if err != nil {
							fmt.Fprintln(connection, "CLIENT_ERROR ", err)
							break
						}

						// Set that we are expecting a value
						client.State = STATE_EXPECTING_VALUE
					} else {
						fmt.Fprintln(connection, "ERROR ", err)

						// Reset the clients state
						client.Reset()
					}
				}
			}

			// Print out errors to stderr
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(connection, "ERROR ", err)
			}
		}(connection)
	}
}

// Reset a clients state to what it would be on first connection
func (client *Client) Reset() {
	client.State = STATE_DEFAULT
	client.Input = []string{}
	client.Command = ""
	client.Record =  &Record{}
}
