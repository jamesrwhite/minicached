# minicache

minicache is a work in progress in-memory caching system
featuring a similar [text based protocol](https://github.com/memcached/memcached/blob/master/doc/protocol.txt) to [memcached](http://memcached.org/) and should be usable from most existing memcached libraries

## Commands implemented

- get
- set
- delete
- flush_all
- quit

## Demo

`go run minicache.go`

````
» telnet localhost 5268
Trying ::1...
Connected to localhost.
Escape character is '^]'.
get a
END
set a 0 0 4
test
STORED
get a
VALUE a 0 4
test
END
````
