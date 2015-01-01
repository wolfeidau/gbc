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

// BConnListener wraps a net Listener and provides BConn rather than net.Conn via accept callback
type BConnListener struct {
	net.Listener
	bufioReaderPool sync.Pool
	bufioWriterPool sync.Pool
	bsize           int
}

// WrapListener builds a new wrapped listener
func WrapListener(l net.Listener) net.Listener {
	return WrapListenerSize(l, defaultBufferSize)
}

// WrapListenerSize builds a new wrapped listener with buffers configured to the specified size
func WrapListenerSize(l net.Listener, bsize int) net.Listener {
	return &BConnListener{Listener: l, bsize: bsize}
}

// Accept accept
func (bcl *BConnListener) Accept() (c net.Conn, err error) {
	c, err = bcl.Listener.Accept()

	if err != nil {
		return
	}

	c = newBConn(c, bcl)

	return
}

// BConn simple buffered connection
type BConn struct {
	net.Conn                   // the underlying net connection
	bufwr    *bufio.ReadWriter // buffered reading/writing from rwc
	bcl      *BConnListener    // the listener who is managing the buffer pools
}

func newBConn(rwc net.Conn, bcl *BConnListener) *BConn {
	c := &BConn{Conn: rwc, bcl: bcl}

	br := c.bcl.newBufioReader(c.Conn)
	bw := c.bcl.newBufioWriter(c.Conn)
	c.bufwr = bufio.NewReadWriter(br, bw)

	return c
}

// ReadWriter access the read writer for this connection
func (c *BConn) ReadWriter() *bufio.ReadWriter {
	return c.bufwr
}

func (c *BConn) finalFlush() {
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
func (c *BConn) Close() (err error) {
	log.Debugf("closing %s", c.RemoteAddr())
	c.finalFlush()
	if c.Conn != nil {
		err = c.Conn.Close()
		c.Conn = nil
	}
	return
}

func (bcl *BConnListener) newBufioReader(r io.Reader) *bufio.Reader {
	if v := bcl.bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReaderSize(r, bcl.bsize)
}

func (bcl *BConnListener) putBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	bcl.bufioReaderPool.Put(br)
}

func (bcl *BConnListener) newBufioWriter(w io.Writer) *bufio.Writer {
	if v := bcl.bufioWriterPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriterSize(w, bcl.bsize)
}

func (bcl *BConnListener) putBufioWriter(bw *bufio.Writer) {
	bw.Reset(nil)
	bcl.bufioWriterPool.Put(bw)
}
