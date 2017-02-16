// minicached is a work in progress in-memory caching system
// featuring a similar text based protocol to memcached
package main

import (
	"github.com/jamesrwhite/minicached/server"
	log "github.com/Sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	server.Listen(5268)
}
