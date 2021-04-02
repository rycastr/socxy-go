package proxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
)

func Handle(newConn net.Conn) *Conn {
	cert := NewCert(2048, 365, "socxy.cloud")
	return handle(newConn, cert)
}

func handle(nConn net.Conn, cert *certificate) *Conn {
	left, right := net.Pipe()

	src := &Conn{
		conn:       right,
		localAddr:  nConn.LocalAddr(),
		remoteAddr: nConn.RemoteAddr(),
	}

	var firstBuf []byte

	firstBuf, src.isTLS = handleConn(nConn)
	if src.isTLS {
		cert, err := tls.X509KeyPair(cert.Certificate(), cert.PrivateKey())
		if err != nil {
			log.Fatal(err)
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		src.conn = tls.Server(right, config)
	}
	go pipe(nConn, left, firstBuf)
	return src
}

func handleConn(rw io.ReadWriter) ([]byte, bool) {
	b, _ := extractBuffer(rw)

	switch {
	case bytes.Equal(b[0:4], []byte("SSH-")):
		return b, false
	case bytes.Equal(b[0:3], []byte{22, 3, 1}):
		return b, true
	default:
		req := bytes.Split(b, []byte("\r\n\r\n"))
		if bytes.Equal(req[1][0:4], []byte("SSH-")) {
			rw.Write([]byte("HTTP/1.1 200\r\n\r\n"))
			return req[1], false
		}
		rw.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return handleConn(rw)
	}
}

func extractBuffer(r io.Reader) ([]byte, error) {
	buffer := make([]byte, 32*1024)
	readed, err := r.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[0:readed], nil
}

func pipe(left, right net.Conn, fb []byte) {
	defer left.Close()
	defer right.Close()

	right.Write(fb)

	go func() {
		io.Copy(left, right)
		left.SetReadDeadline(time.Now().Add(5 * time.Second))
	}()

	io.Copy(right, left)
	right.SetReadDeadline(time.Now().Add(5 * time.Second))
}
