# minicache

minicache is a work in progress in-memory caching system
featuring a similar text based protocol to memcached/redis

## Demo

`go run minicache.go`

````
» telnet localhost 5268
Trying ::1...
Connected to localhost.
Escape character is '^]'.
get test
VALUE get 0 4
test
END
````