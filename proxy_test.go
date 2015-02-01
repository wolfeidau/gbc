package gbc

import (
	"net"
	"testing"
	"time"

	"github.com/juju/loggo"
)

var pconnTests = []struct {
	net  string
	addr string
}{
	{"tcp", "127.0.0.1:0"},
}

const psomeTimeout = 1 * time.Second

func TestPConnCreate(t *testing.T) {
	for _, tt := range pconnTests {

		ln, err := Listen(tt.net, tt.addr)
		if err != nil {
			t.Fatalf("Listen failed: %v", err)
		}
		defer ln.Close()

		loggo.GetLogger("").SetLogLevel(loggo.DEBUG)

		SetLogger(loggo.GetLogger("proxy-test"))

		pln := &ProxyListener{ln}

		errchan := make(chan error)
		donechan := make(chan bool)

		go func() {
			c, err := pln.Accept()
			if err != nil {
				errchan <- err
				return
			}
			defer c.Close()

			// cast the connection
			pconn := c.(*ProxyConn)

			log.Debugf("local %+v", pconn.LocalAddr())

			log.Debugf("remote %+v", pconn.RemoteAddr())

			// check the read writer is created
			if pconn.RemoteAddr().String() != "10.1.1.1:1000" {
				t.Fatal("RemoteAddr is incorrect")
			}

			rb := make([]byte, 128)
			if _, err := pconn.ReadWriter().Read(rb); err != nil {
				errchan <- err
				return
			}

			donechan <- true

		}()

		c, err := net.Dial(tt.net, ln.Addr().String())
		if err != nil {

		}
		defer c.Close()

		c.SetDeadline(time.Now().Add(psomeTimeout))
		c.SetReadDeadline(time.Now().Add(psomeTimeout))
		c.SetWriteDeadline(time.Now().Add(psomeTimeout))

		if _, err := c.Write([]byte("PROXY TCP4 10.1.1.1 20.2.2.2 1000 2000\r\n")); err != nil {
			t.Fatalf("Conn.Write failed: %v", err)
		}

		if _, err := c.Write([]byte("HELLO")); err != nil {
			t.Fatalf("Conn.Write failed: %v", err)
		}

		select {
		case done := <-donechan:
			log.Debugf("done: %v", done)
		case err := <-errchan:
			t.Fatalf("test failed: %v", err)
		}

	}
}
