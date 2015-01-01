package gbc

import (
	"net"
	"testing"
	"time"
)

var connTests = []struct {
	net  string
	addr string
}{
	{"tcp", "127.0.0.1:0"},
}

const someTimeout = 10 * time.Second

func TestBConnCreate(t *testing.T) {
	for _, tt := range connTests {

		ln, err := Listen(tt.net, tt.addr)
		if err != nil {
			t.Fatalf("Listen failed: %v", err)
		}
		defer ln.Close()

		done := make(chan int)

		go func() {
			c, err := ln.Accept()
			if err != nil {
				t.Fatalf("Accept failed: %v", err)
			}
			defer c.Close()

			// cast the connection
			bconn := c.(*BufferedConn)

			// check the read writer is created
			if bconn.bufwr == nil {
				t.Fatal("ReadWriter is nil")
			}

			rb := make([]byte, 128)
			if _, err := bconn.ReadWriter().Read(rb); err != nil {
				t.Fatalf("Conn.Read failed: %v", err)
			}

			done <- 1

		}()

		c, err := net.Dial(tt.net, ln.Addr().String())
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}
		defer c.Close()

		c.SetDeadline(time.Now().Add(someTimeout))
		c.SetReadDeadline(time.Now().Add(someTimeout))
		c.SetWriteDeadline(time.Now().Add(someTimeout))

		if _, err := c.Write([]byte("CONN TEST")); err != nil {
			t.Fatalf("Conn.Write failed: %v", err)
		}

		<-done
	}
}
