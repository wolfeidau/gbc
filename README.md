# gbc [![GoDoc](https://img.shields.io/badge/godoc-Reference-brightgreen.svg?style=flat)](http://godoc.org/github.com/wolfeidau/gbc)

This library provides a wrapper for net.Listener which returns a buffered `net.Conn` enabling you 
to avoid writing another buffered connection wrapper in your service. 

It uses sync.Pool to reuse buffered readers and writers, which are allocated using the default size of `1024`, or the 
value provided to `WrapListenerSize`.

# usage

```go
import "github.com/wolfeidau/gbc"


func main() {
	
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}

	ln = gbc.WrapListener(ln)

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

	// use the rw to do peak/flush ect
	...

}

```

# License

This code is Copyright (c) 2014 Mark Wolfe and licenced under the MIT licence. All rights not explicitly granted in the MIT license are reserved. See the included LICENSE.md file for more details.