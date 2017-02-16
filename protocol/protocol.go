package protocol

import (
	"github.com/jamesrwhite/minicached/server"
	storeInterface "github.com/jamesrwhite/minicached/store"
)

type Protocol interface {
	Process(client *server.Client, datastore *store.Store)
}
