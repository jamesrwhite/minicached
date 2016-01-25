// minicached is a work in progress in-memory caching system
// featuring a similar text based protocol to memcached
package main

// TODO:
// - TTL handling
// - casunique?
// - noreply?
// - Investigate how memcached handles concurrency/locking

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
	Id         string
	Connection net.Conn
	State      uint8
	Input      []string
	Command    string
	Record     *Record
}

type Record struct {
	Key    string
	Value  []byte
	Flags  int64
	Ttl    int64
	Length int64
}

const STATE_DEFAULT uint8 = 1
const STATE_EXPECTING_VALUE uint8 = 2

const STATE_COMMAND_GET uint8 = 3
const STATE_COMMAND_SET uint8 = 4
const STATE_COMMAND_DELETE uint8 = 5
const STATE_COMMAND_QUIT uint8 = 6
const STATE_COMMAND_FLUSH_ALL uint8 = 7

var ticker = time.NewTicker(time.Second * 1)
var clients map[string]*Client
var datastore map[string]*Record

func main() {
	// Start the server on port 5268
	server, err := net.Listen("tcp", ":5268")

	if err != nil {
		log.Fatal(err)
	}

	// Ensure that the server closes
	defer server.Close()

	// Our datastore..
	datastore = make(map[string]*Record)

	// A list of active clients
	clients = make(map[string]*Client)

	// Print out the datastore contents every second for debug
	go func() {
		for range ticker.C {
			fmt.Println(datastore)
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
		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {
	// Create the client and store it
	client := &Client{
		Id:         connection.RemoteAddr().String(),
		Connection: connection,
		State:      STATE_DEFAULT,
		Input:      []string{},
		Command:    "",
		Record:     &Record{},
	}
	clients[client.Id] = client

	// Ensure the client is tidied up once they're done
	defer func(client *Client, clients map[string]*Client) {
		delete(clients, client.Id)
		client.Connection.Close()
	}(client, clients)

	// Create a new scanner for the client input
	scanner := bufio.NewScanner(client.Connection)

	fmt.Println("[CONNECTION] new connection: ", client.Connection)

	// Handle each line (command)
	for scanner.Scan() {
		// Split the client input up based on spaces
		client.Input = strings.Split(scanner.Text(), " ")

		fmt.Println("[CLIENT] input received: ", client.Input)

		// Set the clients state
		err := setClientState(client)

		// If there was an erorr setting the client state skip and
		// wait for more input
		if err != nil {
			continue
		}

		// Switch on the type of command
		switch client.State {
		// Are we expecting a value from a set command?
		case STATE_EXPECTING_VALUE:
			handleStateExpectingValue(client, scanner.Bytes())
		// get [key1 ... keyn]
		// TODO: handling multiple key gets
		case STATE_COMMAND_GET:
			handleStateCommandGet(client)
		// set [key] [flags] [exptime] [length] [casunique] [noreply]
		case STATE_COMMAND_SET:
			err := handleStateCommandSet(client)

			if err != nil {
				break
			}
		// delete [key] [noreply]
		case STATE_COMMAND_DELETE:
			handleStateCommandDelete(client)
		// quit
		case STATE_COMMAND_QUIT:
			handleStateCommandQuit(client)
		// flush_all [delay] [noreply]
		// TODO: handle noreply
		case STATE_COMMAND_FLUSH_ALL:
			err := handleStateCommandFlushAll(client)

			if err != nil {
				break
			}
		}
	}

	// Print out errors to stderr
	if err := scanner.Err(); err != nil {
		fmt.Println("[SERVER] error ", err)
		fmt.Fprintln(client.Connection, "ERROR ", err)
		client.Reset()
	}
}

func setClientState(client *Client) error {
	fmt.Println("[CLIENT] pre parse state: ", client.State)

	// Determine the clients state based on the command unless
	// we're waiting for a value
	if client.State != STATE_EXPECTING_VALUE {
		// Get the command
		client.Command = client.Input[0]

		fmt.Println("[CLIENT] setting command: ", client.Command)

		switch client.Command {
		case "get":
			client.State = STATE_COMMAND_GET
		case "set":
			client.State = STATE_COMMAND_SET
		case "delete":
			client.State = STATE_COMMAND_DELETE
		case "flush_all":
			client.State = STATE_COMMAND_FLUSH_ALL
		case "quit":
			client.State = STATE_COMMAND_QUIT
		default:
			fmt.Println("[CLIENT] unknown state: ", client.State)
			fmt.Fprintln(client.Connection, "ERROR")
			client.Reset()

			return fmt.Errorf("Unknown client state: %s", client.State)
		}

		fmt.Println("[CLIENT] post parse state: ", client.State)
	}

	return nil
}

func handleStateExpectingValue(client *Client, bytes []byte) {
	// If the value isn't set then set it
	if len(client.Record.Value) == 0 {
		client.Record.Value = bytes
		// Otherwise append to it
	} else {
		client.Record.Value = append(client.Record.Value, bytes...)
	}

	client.Record.Value = append(client.Record.Value, []byte{'\r', '\n'}...)

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
			fmt.Fprintln(client.Connection, "STORED")
			// Otherwise the client has messed up
		} else {
			// Inform the client that they messed up
			fmt.Fprintln(client.Connection, "CLIENT_ERROR")
			fmt.Fprintln(client.Connection, valueLength)
			fmt.Fprintln(client.Connection, "ERROR")
		}

		// Reset the clients state
		client.Reset()
	}
}

func handleStateCommandGet(client *Client) {
	// Check if a key was passed, if so try and retrieve it
	if len(client.Input) == 2 {
		// Get the key
		key := client.Input[1]

		// Look up the record in our datastore
		record := datastore[key]

		// Did it exist?
		if record != nil {
			fmt.Fprintln(client.Connection, fmt.Sprintf("VALUE %s %d %d", record.Key, record.Flags, record.Length))
			fmt.Fprint(client.Connection, string(record.Value[:]))
		}

		fmt.Fprintln(client.Connection, "END")
	} else {
		fmt.Fprintln(client.Connection, "CLIENT_ERROR")
	}

	client.Reset()
}

func handleStateCommandSet(client *Client) error {
	// Check the right number of arguments are passed
	// casunique and noreply are optional
	if len(client.Input) == 5 {
		var err error

		// Get the key name
		client.Record.Key = client.Input[1]

		// Get any flags
		client.Record.Flags, err = strconv.ParseInt(client.Input[2], 10, 64)

		if err != nil {
			fmt.Fprintln(client.Connection, "CLIENT_ERROR ", err)
			client.Reset()

			return fmt.Errorf("Error parsing client command set input: %s", err)
		}

		// Get the key TTL
		client.Record.Ttl, err = strconv.ParseInt(client.Input[3], 10, 64)

		if err != nil {
			fmt.Fprintln(client.Connection, "CLIENT_ERROR ", err)
			client.Reset()

			return fmt.Errorf("Error parsing client command set input: %s", err)
		}

		// Get the value length
		client.Record.Length, err = strconv.ParseInt(client.Input[4], 10, 64)

		if err != nil {
			fmt.Fprintln(client.Connection, "CLIENT_ERROR ", err)
			client.Reset()

			return fmt.Errorf("Error parsing client command set params: %s", err)
		}

		// Set that we are expecting a value
		client.State = STATE_EXPECTING_VALUE
	} else {
		fmt.Fprintln(client.Connection, "CLIENT_ERROR")
		client.Reset()

		return fmt.Errorf("Not enough params sent to command set, expected 5 got: %d", len(client.Input))
	}

	return nil
}

func handleStateCommandDelete(client *Client) {
	// Check if a key was passed, if so try and retrieve it
	if len(client.Input) == 2 {
		// Get the key
		key := client.Input[1]

		// Look up the record in our datastore
		record := datastore[key]

		// Did it exist? If so 'delete' it
		if record != nil {
			delete(datastore, key)
			fmt.Fprintln(client.Connection, "DELETED")
		} else {
			fmt.Fprintln(client.Connection, "NOT_FOUND")
		}

	} else {
		fmt.Fprintln(client.Connection, "CLIENT_ERROR")
	}

	client.Reset()
}

func handleStateCommandQuit(client *Client) {
	// Not much to do here atm..
	// Eventually we will do logging etc
}

func handleStateCommandFlushAll(client *Client) error {
	// Check if a delay was passed
	if len(client.Input) == 2 && client.Input[1] != "" {
		delay, err := strconv.ParseInt(client.Input[1], 10, 64)

		if err != nil {
			fmt.Fprintln(client.Connection, "CLIENT_ERROR ", err)
			client.Reset()

			return fmt.Errorf("Error parsing client command flush_all delay: %s", err)
		}

		time.Sleep(time.Duration(delay) * time.Second)
	}

	// Reset the datastore
	datastore = make(map[string]*Record)
	fmt.Fprintln(client.Connection, "OK")

	client.Reset()

	return nil
}

// Reset a clients state to what it would be on first connection
func (client *Client) Reset() {
	fmt.Println("[CLIENT] resetting client")

	client.State = STATE_DEFAULT
	client.Input = []string{}
	client.Command = ""
	client.Record = &Record{}
}
