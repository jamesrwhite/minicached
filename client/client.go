package client

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/jamesrwhite/minicached/store"
)

type Client struct {
	Id         string
	Connection net.Conn
	State      uint8
	Input      string
	Command    string
	Record     *store.Record
}

const STATE_DEFAULT uint8 = 1
const STATE_EXPECTING_VALUE uint8 = 2
const STATE_COMMAND_GET uint8 = 3
const STATE_COMMAND_SET uint8 = 4
const STATE_COMMAND_DELETE uint8 = 5
const STATE_COMMAND_QUIT uint8 = 6
const STATE_COMMAND_FLUSH_ALL uint8 = 7

// Reset a clients state to what it would be on first connection
// TODO: move to client package
func (client *Client) Reset() {
	log.WithFields(log.Fields{
		"event":         "client_reset",
		"client_state":  client.State,
		"connection_id": client.Connection.RemoteAddr().String(),
	}).Debug("Client Reset")

	client.State = STATE_DEFAULT
	client.Input = ""
	client.Command = ""
	client.Record = &store.Record{}
}
