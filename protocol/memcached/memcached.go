package memcached

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jamesrwhite/minicached/client"
	"github.com/jamesrwhite/minicached/store/memory"
)

func init() {
	log.Info("Initialising memcached protocol parser")
}

func Process(serverClient *client.Client) (response string, err error) {
	// Determine the clients state based on the command unless
	// we're in the state of waiting for a value
	if serverClient.State != client.STATE_EXPECTING_VALUE {
		// Set the command the client should be processing
		serverClient.Command, err = getCommand(serverClient.Input)

		if err != nil {
			return "", err
		}

		log.WithFields(log.Fields{
			"event":         fmt.Sprintf("client_command_%s", serverClient.Command),
			"client_state":  serverClient.State,
			"connection_id": serverClient.Connection.RemoteAddr().String(),
		}).Debug("Client Command Received")

		// Set the clients state based on this command
		serverClient.State, err = getState(serverClient.Command)

		log.WithFields(log.Fields{
			"event":         "client_state_change",
			"client_state":  serverClient.State,
			"connection_id": serverClient.Connection.RemoteAddr().String(),
		}).Debug("Client State Change")

		if err != nil {
			return "", err
		}
	}

	// Dispatch the request based on the clients current state
	return dispatch(serverClient)
}

func getCommand(input string) (string, error) {
	// Split the input string into a slice by spaces
	inputSlice := strings.Split(input, " ")

	if len(inputSlice) < 1 {
		return "", fmt.Errorf("Unable to determine command")
	}

	return inputSlice[0], nil
}

func getState(command string) (uint8, error) {
	switch command {
	case "get":
		return client.STATE_COMMAND_GET, nil
	case "set":
		return client.STATE_COMMAND_SET, nil
	case "delete":
		return client.STATE_COMMAND_DELETE, nil
	case "flush_all":
		return client.STATE_COMMAND_FLUSH_ALL, nil
	case "quit":
		return client.STATE_COMMAND_QUIT, nil
	default:
		return 0, fmt.Errorf("ERROR")
	}
}

func dispatch(serverClient *client.Client) (string, error) {
	// Switch on the type of command
	switch serverClient.State {
	// Are we expecting a value from a set command?
	case client.STATE_EXPECTING_VALUE:
		return processStateExpectingValue(serverClient)
	// get [key1 ... keyn]
	// TODO: handling multiple key gets
	case client.STATE_COMMAND_GET:
		return processStateCommandGet(serverClient)
	// set [key] [flags] [exptime] [length] [casunique] [noreply]
	case client.STATE_COMMAND_SET:
		return processStateCommandSet(serverClient)
	// delete [key] [noreply]
	case client.STATE_COMMAND_DELETE:
		return processStateCommandDelete(serverClient)
	// quit
	case client.STATE_COMMAND_QUIT:
		return processStateCommandQuit(serverClient)
	// flush_all [delay] [noreply]
	// TODO: handle noreply
	case client.STATE_COMMAND_FLUSH_ALL:
		return processStateCommandFlushAll(serverClient)
	}

	return "", fmt.Errorf("Unknown state %d", serverClient.State)
}

func processStateExpectingValue(serverClient *client.Client) (string, error) {
	// If the value isn't set then set it
	if len(serverClient.Record.Value) == 0 {
		serverClient.Record.Value = serverClient.Input
	} else {
		serverClient.Record.Value += serverClient.Input
	}

	serverClient.Record.Value += "\r\n"

	// Count the length of the value minus the trailing \r\n
	valueLength := int64(len(serverClient.Record.Value)) - 2

	// If the datastore is same or greater than the expected length
	// we are done with this op
	if valueLength >= serverClient.Record.Length {
		// If it's the same length we can try and store it
		if valueLength == serverClient.Record.Length {
			memory.Set(serverClient.Record.Key, serverClient.Record.Value, serverClient.Record.Length, serverClient.Record.Flags, serverClient.Record.Ttl)

			// Reset the clients state as we are no longer expecting a value
			serverClient.Reset()

			return "STORED\n", nil
		}

		// Reset the clients state as we are no longer expecting a value
		serverClient.Reset()

		// Inform the client that they messed up
		return "", fmt.Errorf(fmt.Sprintf("CLIENT_ERROR\n%d\nERROR", valueLength))
	}

	return "", nil
}

func processStateCommandGet(serverClient *client.Client) (string, error) {
	// Split the input string into a slice by spaces
	inputSlice := strings.Split(serverClient.Input, " ")

	// Check if a key was passed, if so try and retrieve it
	if len(inputSlice) == 2 {
		// Get the key
		key := inputSlice[1]

		// Look up the record in our datastore
		found, record := memory.Get(key)

		var response string

		// Did it exist?
		if found == true {
			response = fmt.Sprintf("VALUE %s %d %d\n%s", record.Key, record.Flags, record.Length, record.Value)
		}

		response += "END\n"

		return response, nil
	}

	return "", fmt.Errorf("CLIENT_ERROR")
}

func processStateCommandSet(serverClient *client.Client) (string, error) {
	// Split the input string into a slice by spaces
	inputSlice := strings.Split(serverClient.Input, " ")

	// Check the right number of arguments are passed
	// casunique and noreply are optional
	if len(inputSlice) == 5 {
		var err error

		// Get the key name
		serverClient.Record.Key = inputSlice[1]

		// Get any flags
		serverClient.Record.Flags, err = strconv.ParseInt(inputSlice[2], 10, 64)

		if err != nil {
			return "", fmt.Errorf("CLIENT_ERROR")
		}

		// Get the key TTL
		serverClient.Record.Ttl, err = strconv.ParseInt(inputSlice[3], 10, 64)

		if err != nil {
			return "", fmt.Errorf("CLIENT_ERROR")
		}

		// Get the value length
		serverClient.Record.Length, err = strconv.ParseInt(inputSlice[4], 10, 64)

		if err != nil {
			return "", fmt.Errorf("CLIENT_ERROR")
		}

		// Set that we are expecting a value
		serverClient.State = client.STATE_EXPECTING_VALUE

		log.WithFields(log.Fields{
			"event":         "client_state_change",
			"client_state":  serverClient.State,
			"connection_id": serverClient.Connection.RemoteAddr().String(),
		}).Debug("Client State Change")
	} else {
		return "", fmt.Errorf("CLIENT_ERROR")
	}

	return "", nil
}

func processStateCommandDelete(serverClient *client.Client) (string, error) {
	// Split the input string into a slice by spaces
	inputSlice := strings.Split(serverClient.Input, " ")

	// Check if a key was passed, if so try and retrieve it
	if len(inputSlice) == 2 {
		// Get the key
		key := inputSlice[1]

		// Look up the record in our datastore
		found, _ := memory.Get(key)

		// Did it exist? If so 'delete' it
		if found == true {
			memory.Delete(key)

			return "DELETED\n", nil
		} else {
			return "NOT_FOUND\n", nil
		}

	}

	return "", fmt.Errorf("CLIENT_ERROR")
}

func processStateCommandQuit(serverClient *client.Client) (string, error) {
	// Not much to do here atm..
	// Eventually we will do logging etc

	return "\n", nil
}

func processStateCommandFlushAll(serverClient *client.Client) (string, error) {
	// Split the input string into a slice by spaces
	inputSlice := strings.Split(serverClient.Input, " ")

	// Check if a delay was passed
	if len(inputSlice) == 2 && inputSlice[1] != "" {
		delay, err := strconv.ParseInt(inputSlice[1], 10, 64)

		if err != nil {
			return "", fmt.Errorf("CLIENT_ERROR")
		}

		time.Sleep(time.Duration(delay) * time.Second)
	}

	// Reset the datastore
	memory.Flush()

	return "OK\n", nil
}
