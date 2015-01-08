# gbc [![GoDoc](https://img.shields.io/badge/godoc-Reference-brightgreen.svg?style=flat)](http://godoc.org/github.com/wolfeidau/gbc)

This library provides a wrapper for net.Listener which returns `net.Conn` with buffered preconfigured enabling you 
to avoid writing another connection wrapper with buffers attached. 

The aim are:

* provide a basis for layered listeners which can share the same buffers
* provide some wrappers for the buffered listener which perform various functions prior to calling `Accept`, such as vetting connections, or read preamble such as the proxy protocol example.

It uses sync.Pool to reuse buffered readers and writers, which are allocated using the default size of `1024`, or the 
value provided to `SetBufferSize`.

# usage

```go
import (
	"log"

	"github.com/wolfeidau/gbc"
)


func main() {
	
	ln, err := gbc.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}

	// wrap the buffered listener in the proxy header listener
	// see http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt 
	// this is used by haproxy and Elastic Load Balancer to pass through 
	// source and destination ip addresses to TCP listeners behind these services.
	pln := &gbc.ProxyListener{ln}

	for {
		conn, err := pln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	defer conn.Close() // this will flush if needed

	// print the remote address
	log.Printf("remote addr: %s", conn.RemoteAddr().String())

	// cast the connection to the buffered connection
	bconn := conn.(*gbc.BConn)

	// get the read writer for this buffered connection
	rw := bconn.ReadWriter()

	// use rw to do peak/flush ect
	...

}

```

# References

* [Proxy Protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt)
* [go-proxyproto](https://github.com/armon/go-proxyproto) borrowed the initial proxy protocol parsing from this library!
* [Enabling Proxy Protocol on ELB](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html)
* [Elastic Load Balancing adds Support for Proxy Protocol](https://aws.amazon.com/blogs/aws/elastic-load-balancing-adds-support-for-proxy-protocol/)

# Sponsor

This project was made possible by [Ninja Blocks](http://ninjablocks.com).

# License

This code is Copyright (c) 2014 Mark Wolfe and licenced under the MIT licence. All rights not explicitly granted in the MIT license are reserved. See the included LICENSE.md file for more details.