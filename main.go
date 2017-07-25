// minicached is a work in progress in-memory caching system
// featuring a similar text based protocol to memcached
package main

import (
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/jamesrwhite/minicached/server"
)

var port = 5268

func main() {
	log.SetLevel(log.DebugLevel)

	portString := os.Getenv("MINICACHED_PORT")

	if portString == "" {
		log.Infof("MINICACHED_PORT not set, defaulting to %d", port)
	} else {
		var err error
		port, err = strconv.Atoi(portString)

		if err != nil {
			log.Fatal("Unable to parse port from MINICACHED_PORT")
		}
	}

	server.Listen(port)
}
