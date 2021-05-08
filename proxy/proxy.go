package proxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
)

func Handle(newConn net.Conn, cert *Certificate) (*Conn, error) {
	return handle(newConn, cert)
}

func handle(nConn net.Conn, cert *Certificate) (*Conn, error) {
	left, right := net.Pipe()

	src := &Conn{
		conn:       right,
		localAddr:  nConn.LocalAddr(),
		remoteAddr: nConn.RemoteAddr(),
	}

	var (
		firstBuf []byte
		err      error
	)

	firstBuf, src.isTLS, err = handleConn(nConn)
	if err != nil {
		nConn.Close()
		return nil, err
	}

	if src.isTLS {
		cert, err := tls.X509KeyPair(cert.Certificate(), cert.PrivateKey())
		if err != nil {
			log.Fatal(err)
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		src.conn = tls.Server(right, config)
	}
	go pipe(nConn, left, firstBuf)
	return src, nil
}

func handleConn(rw io.ReadWriter) ([]byte, bool, error) {
	b, err := extractBuffer(rw)
	if err != nil {
		return nil, false, err
	}

	switch {
	case bytes.Equal(b[0:4], []byte("SSH-")):
		return b, false, nil
	case bytes.Equal(b[0:3], []byte{22, 3, 1}):
		return b, true, nil
	default:
		req := bytes.Split(b, []byte("\r\n\r\n"))
		if bytes.Equal(req[1][0:4], []byte("SSH-")) {
			rw.Write([]byte("HTTP/1.1 200\r\n\r\n"))
			return req[1], false, nil
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

	go func() {
		io.Copy(left, right)
		left.SetReadDeadline(time.Now().Add(5 * time.Second))
	}()

	right.Write(fb)

	io.Copy(right, left)
	right.SetReadDeadline(time.Now().Add(5 * time.Second))
}
