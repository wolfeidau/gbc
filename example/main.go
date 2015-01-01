package main

import (
	"log"
	"net"

	"github.com/wolfeidau/gbc"
)

var (
	listenAddr = ":2020"
)

type logger struct{}

func (logger) Debugf(message string, args ...interface{}) { log.Printf(message, args...) }

func main() {

	gbc.SetLogger(&logger{})

	ln, err := net.Listen("tcp", listenAddr)

	if err != nil {
		// handle error
		log.Fatalf("listen failed - %s", err)
	}

	log.Printf("listening on %s", listenAddr)

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

	defer conn.Close()

	// cast the connection
	bconn := conn.(gbc.BufferedConn)

	rw := bconn.ReadWriter()

	log.Printf("bconn %v", bconn)
	log.Printf("rw %v", rw)

	// use the rw to do peak/flush ect

}
