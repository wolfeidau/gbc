package gbc

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var (
	prefix    = []byte("PROXY ")
	prefixLen = len(prefix)
)

// ProxyListener wrapper for the BufferedConnListener which provides proxy protocol support.
type ProxyListener struct {
	*BufferedConnListener
}

// Accept waits for and returns the next connection to the listener.
func (p *ProxyListener) Accept() (net.Conn, error) {
	// Get the underlying connection
	conn, err := p.BufferedConnListener.Accept()
	if err != nil {
		return nil, err
	}
	// cast the connection
	bconn := conn.(*BufferedConn)

	pconn := &ProxyConn{bconn, nil, nil}

	err = pconn.checkPrefix()

	if err != nil {
		return nil, err
	}

	return pconn, nil
}

// ProxyConn wrapper for BufferedConn which adds dst and src addr fields parsed from the proxy protocol
type ProxyConn struct {
	*BufferedConn
	dstAddr *net.TCPAddr
	srcAddr *net.TCPAddr
}

// RemoteAddr returns the remote network address, if a proxy preamble was present it uses the source address
// provided.
func (p *ProxyConn) RemoteAddr() net.Addr {
	if p.srcAddr != nil {
		return p.srcAddr
	}
	return p.BufferedConn.RemoteAddr()
}

// ProxyLocalAddr This returns the destination address within the proxy header, or
// if not present the local address.
func (p *ProxyConn) ProxyLocalAddr() net.Addr {
	if p.dstAddr != nil {
		return p.dstAddr
	}
	return p.BufferedConn.LocalAddr()
}

func (p *ProxyConn) checkPrefix() error {

	// Incrementally check each byte of the prefix
	for i := 1; i <= prefixLen; i++ {
		inp, err := p.ReadWriter().Peek(i)
		if err != nil {
			return err
		}

		// Check for a prefix mis-match, quit early
		if !bytes.Equal(inp, prefix[:i]) {
			return nil
		}
	}

	// Read the header line
	header, err := p.ReadWriter().ReadString('\n')
	if err != nil {
		p.Close()
		return err
	}

	// Strip the carriage return and new line
	header = header[:len(header)-2]

	// Split on spaces, should be (PROXY <type> <src addr> <dst addr> <src port> <dst port>)
	parts := strings.Split(header, " ")
	if len(parts) != 6 {
		p.Close()
		return fmt.Errorf("Invalid header line: %s", header)
	}

	// Verify the type is known
	switch parts[1] {
	case "TCP4":
	case "TCP6":
	default:
		p.Close()
		return fmt.Errorf("Unhandled address type: %s", parts[1])
	}

	// Parse out the source address
	ip := net.ParseIP(parts[2])
	if ip == nil {
		p.Close()
		return fmt.Errorf("Invalid source ip: %s", parts[2])
	}
	port, err := strconv.Atoi(parts[4])
	if err != nil {
		p.Close()
		return fmt.Errorf("Invalid source port: %s", parts[4])
	}
	p.srcAddr = &net.TCPAddr{IP: ip, Port: port}

	// Parse out the destination address
	ip = net.ParseIP(parts[3])
	if ip == nil {
		p.Close()
		return fmt.Errorf("Invalid destination ip: %s", parts[3])
	}
	port, err = strconv.Atoi(parts[5])
	if err != nil {
		p.Close()
		return fmt.Errorf("Invalid destination port: %s", parts[5])
	}
	p.dstAddr = &net.TCPAddr{IP: ip, Port: port}

	return nil
}
