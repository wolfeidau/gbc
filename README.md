# gbc [![GoDoc](https://img.shields.io/badge/godoc-Reference-brightgreen.svg?style=flat)](http://godoc.org/github.com/wolfeidau/gbc)

This library provides a wrapper for net.Listener which returns `net.Conn` with buffered preconfigured enabling you 
to avoid writing another connection wrapper with buffers attached. 

The aim are:

* provide a basis for layered listeners which can share the same buffers
* provide a callback prior to accept which can vet connections prior to passing onto `Accept`

It uses sync.Pool to reuse buffered readers and writers, which are allocated using the default size of `1024`, or the 
value provided to `SetBufferSize`.

# usage

```go
import "github.com/wolfeidau/gbc"


func main() {
	
	ln, err := gbc.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	defer conn.Close() // this will flush if needed

	// cast the connection to the buffered connection
	bconn := conn.(*gbc.BConn)

	// get the read writer for this buffered connection
	rw := bconn.ReadWriter()

	// use rw to do peak/flush ect
	...

}

```

# License

This code is Copyright (c) 2014 Mark Wolfe and licenced under the MIT licence. All rights not explicitly granted in the MIT license are reserved. See the included LICENSE.md file for more details.