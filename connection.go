package gbc

import (
	"bufio"
	"io"
	"net"
	"sync"
)

// noLimit is an effective infinite upper bound for io.LimitedReader
const (
	defaultBufferSize = 1024
)

var (
	// value used when the next listener is allocated
	bufferSize = defaultBufferSize
)

// SetBufferSize assigs the buffer size used when Listen is next called, this must be done before that happens of course
func SetBufferSize(size int) {
	bufferSize = size
}

type BeforeAccept func(bconn *BufferedConn) error

// BufferedConnListener wraps a net Listener and provides BufferedConn rather than net.Conn via accept callback
type BufferedConnListener struct {
	net.Listener
	bufioReaderPool sync.Pool
	bufioWriterPool sync.Pool
	bsize           int
	before          BeforeAccept
}

// Listen announces on the local network address laddr. The network net must be a stream-oriented
// network: "tcp", "tcp4", "tcp6", "unix" or "unixpacket". See net.Dial for the syntax of laddr.
func Listen(network, laddr string) (net.Listener, error) {
	ln, err := net.Listen(network, laddr)
	if err != nil {
		return ln, err
	}
	return &BufferedConnListener{Listener: ln, bsize: bufferSize}, nil
}

// Accept accept
func (bcl *BufferedConnListener) Accept() (c net.Conn, err error) {
	c, err = bcl.Listener.Accept()

	if err != nil {
		return
	}

	c = newBufferedConn(c, bcl)

	// cast the connection
	bconn := c.(*BufferedConn)

	if bcl.before != nil {
		err = bcl.before(bconn)
	}

	if err != nil {
		c.Close()
		c = nil
	}

	return
}

// SetBeforeAccept assign a before accept function which can intercept, use and reject connections
func (bcl *BufferedConnListener) SetBeforeAccept(before BeforeAccept) {
	bcl.before = before
}

// BufferedConn simple buffered connection
type BufferedConn struct {
	net.Conn                       // the underlying net connection
	bufwr    *bufio.ReadWriter     // buffered reading/writing from rwc
	bcl      *BufferedConnListener // the listener who is managing the buffer pools
}

func newBufferedConn(rwc net.Conn, bcl *BufferedConnListener) *BufferedConn {
	c := &BufferedConn{Conn: rwc, bcl: bcl}

	br := c.bcl.newBufioReader(c.Conn)
	bw := c.bcl.newBufioWriter(c.Conn)
	c.bufwr = bufio.NewReadWriter(br, bw)

	return c
}

// Read read from the underlying buffered readwriter to avoid issues
func (c *BufferedConn) Read(b []byte) (int, error) {
	return c.bufwr.Read(b)
}

// ReadWriter access the read writer for this connection
func (c *BufferedConn) ReadWriter() *bufio.ReadWriter {
	return c.bufwr
}

func (c *BufferedConn) finalFlush() {
	if c.bufwr != nil {
		c.bufwr.Flush()

		// Steal the bufio.Reader (~4KB worth of memory) and its associated
		// reader for a future connection.
		c.bcl.putBufioReader(c.bufwr.Reader)

		// Steal the bufio.Writer (~4KB worth of memory) and its associated
		// writer for a future connection.
		c.bcl.putBufioWriter(c.bufwr.Writer)

		c.bufwr = nil
	}
}

// Close the connection.
func (c *BufferedConn) Close() (err error) {
	log.Debugf("closing %s", c.RemoteAddr())
	c.finalFlush()
	if c.Conn != nil {
		err = c.Conn.Close()
		c.Conn = nil
	}
	return
}

func (bcl *BufferedConnListener) newBufioReader(r io.Reader) *bufio.Reader {
	if v := bcl.bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReaderSize(r, bcl.bsize)
}

func (bcl *BufferedConnListener) putBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	bcl.bufioReaderPool.Put(br)
}

func (bcl *BufferedConnListener) newBufioWriter(w io.Writer) *bufio.Writer {
	if v := bcl.bufioWriterPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriterSize(w, bcl.bsize)
}

func (bcl *BufferedConnListener) putBufioWriter(bw *bufio.Writer) {
	bw.Reset(nil)
	bcl.bufioWriterPool.Put(bw)
}
