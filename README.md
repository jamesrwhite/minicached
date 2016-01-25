# minicached

minicached is a work in progress in-memory caching system
featuring a similar [text based protocol](https://github.com/memcached/memcached/blob/master/doc/protocol.txt)
to [memcached](http://memcached.org/) and should be usable from most existing memcached libraries. The eventual
goal being to have a similar feature set to memcached but also to support prefix based wildcards on keys so you can do things
like delete all keys that start with `user` by doing `delete user*`.

## Commands implemented

- get
- set
- delete
- flush_all
- quit

## Demo

`go run minicached.go`

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

## Tests

- `composer install`
- `make test`

## Build

- `make build`

## Install

- `make install`

## License

The MIT License (MIT)

Copyright (c) 2016 James White

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
