package protocol

import (
	"github.com/jamesrwhite/minicached/server"
	"github.com/jamesrwhite/minicached/store"
)

type Protocol interface {
	Process(client *server.Client, datastore *store.Store)
	Error(err error) string
}
