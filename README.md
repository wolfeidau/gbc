# gbc

This library provides a wrapper for net.Listener which returns buffered wrapper around `net.Conn` 
which enables you to avoid writing another buffered connection wrapper in your service. 

It uses sync.Pool to reuse buffers, these are allocated using the default size of `1024`, or the 
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

# API

```go
import "github.com/wolfeidau/gbc"
```

## func SetLogger
``` go
func SetLogger(l Logger)
```
SetLogger assign a logger for debugging gbc


## func WrapListener
``` go
func WrapListener(l net.Listener) net.Listener
```
WrapListener builds a new wrapped listener


## func WrapListenerSize
``` go
func WrapListenerSize(l net.Listener, bsize int) net.Listener
```
WrapListenerSize builds a new wrapped listener with buffers configured to the specified size



## type BConn
``` go
type BConn struct {
    net.Conn // the underlying net connection
    // contains filtered or unexported fields
}
```
BConn simple buffered connection

### func (\*BConn) Close
``` go
func (c *BConn) Close() (err error)
```
Close the connection.

### func (\*BConn) ReadWriter
``` go
func (c *BConn) ReadWriter() *bufio.ReadWriter
```
ReadWriter access the read writer for this connection

## type BConnListener
``` go
type BConnListener struct {
    net.Listener
    // contains filtered or unexported fields
}
```
BConnListener wraps a net Listener and provides BConn rather than net.Conn via accept callback


### func (\*BConnListener) Accept
``` go
func (bcl *BConnListener) Accept() (c net.Conn, err error)
```
Accept accept



## type Logger
``` go
type Logger interface {
    Debugf(message string, args ...interface{})
}
```
Logger used to enable debugging


# Author

Mark Wolfe mark@wolfe.id.au

# License

This code is Copyright (c) 2014 Mark Wolfe and licenced under the MIT licence. All rights not explicitly granted in the MIT license are reserved. See the included LICENSE.md file for more details.

- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)