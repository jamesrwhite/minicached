package server

// TODO:
// - TTL handling
// - casunique?
// - noreply?
// - Investigate how memcached handles concurrency/locking

import (
	"bufio"
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/jamesrwhite/minicached/client"
	"github.com/jamesrwhite/minicached/protocol/memcached"
	"github.com/jamesrwhite/minicached/store"
)

var serverClients map[string]*client.Client

func Listen(port int) {
	// Initialise the store
	log.Info("Initialising memory store")

	log.Infof("Listening on port %d", port)
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		log.Fatal(err)
	}

	// Ensure that the server closes
	defer server.Close()

	// A list of active clients
	serverClients = make(map[string]*client.Client)

	for {
		// Wait for a connection.
		connection, err := server.Accept()

		if err != nil {
			log.Fatal(err)
		}

		log.WithFields(log.Fields{
			"event":         "connection_open",
			"connection_id": connection.RemoteAddr().String(),
		}).Debug("Connection opened")

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting new connections,
		// so that multiple connections may be served concurrently.
		go process(connection)
	}
}

func process(connection net.Conn) {
	// Create the client and store it
	serverClient := &client.Client{
		Id:         connection.RemoteAddr().String(),
		Connection: connection,
		State:      client.STATE_DEFAULT,
		Input:      "",
		Command:    "",
		Record:     &store.Record{},
	}

	// TODO: locking
	serverClients[serverClient.Id] = serverClient

	log.WithFields(log.Fields{
		"event":         "client_create",
		"client_state":  serverClient.State,
		"connection_id": serverClient.Id,
	}).Debug("Client created")

	// Ensure the client is tidied up once they're done
	defer func(serverClient *client.Client, serverClients map[string]*client.Client) {
		// Delete the client from the clients map
		log.WithFields(log.Fields{
			"event":         "client_delete",
			"client_state":  serverClient.State,
			"connection_id": serverClient.Id,
		}).Debug("Client delete")

		delete(serverClients, serverClient.Id)

		// Close the clients connection
		log.WithFields(log.Fields{
			"event":         "connection_close",
			"client_state":  serverClient.State,
			"connection_id": serverClient.Id,
		}).Debug("Connection closed")

		serverClient.Connection.Close()
	}(serverClient, serverClients)

	// Create a new scanner for the client input
	scanner := bufio.NewScanner(serverClient.Connection)

	// Handle each line (command)
	for scanner.Scan() {
		// Get the raw text input from the client
		scannedText := scanner.Text()
		log.WithFields(log.Fields{
			"event":         "client_input",
			"client_state":  serverClient.State,
			"connection_id": serverClient.Id,
		}).Debug(fmt.Sprintf("Client Input: %s", scannedText))

		// Set the client input to be the text we scanned in
		serverClient.Input = scannedText

		// Process the command based on the clients input
		response, err := memcached.Process(serverClient)

		if err != nil {
			fmt.Fprintln(serverClient.Connection, err)

			// An error occured so reset the client state
			serverClient.Reset()

			log.WithFields(log.Fields{
				"event":         "client_error",
				"client_state":  serverClient.State,
				"connection_id": serverClient.Id,
			}).Debug(fmt.Sprintf("Client Error: %s", err.Error()))
		} else {
			fmt.Fprint(serverClient.Connection, response)
		}
	}

	// Print out errors to stderr
	if err := scanner.Err(); err != nil {
		log.WithFields(log.Fields{
			"event":         "client_error",
			"client_state":  serverClient.State,
			"connection_id": serverClient.Id,
		}).Debug(fmt.Sprintf("Client Scan Error: %s", err.Error()))

		fmt.Fprintln(serverClient.Connection, "ERROR ", err)

		serverClient.Reset()
	}
}
