// minicached is a work in progress in-memory caching system
// featuring a similar text based protocol to memcached
package main

import (
	"github.com/jamesrwhite/minicached/server"
)

func main() {
	server.Listen(5268)
}
