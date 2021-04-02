package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/socxy/socxy-server/proxy"
	gossh "golang.org/x/crypto/ssh"
)

func connStatus(ctx ssh.Context) {
	<-ctx.Done()
	log.Printf("Connection %s closed by: %s", ctx.RemoteAddr().String(), ctx.User())
}

func listen(port int) {
	addr := &net.TCPAddr{Port: port}
	server := &ssh.Server{
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			log.Printf("New connection %s on port %s, secure(TLS/SSL): %v, client: %s, authenticated by: %s\n",
				ctx.RemoteAddr().String(), strings.Split(ctx.LocalAddr().String(), ":")[1],
				ctx.Value("tls"), ctx.ClientVersion(), ctx.User())
			go connStatus(ctx)
			return true
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"direct-tcpip": ssh.DirectTCPIPHandler,
		},
		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		ConnCallback: func(ctx ssh.Context, conn net.Conn) net.Conn {
			c := proxy.Handle(conn)
			ctx.SetValue("tls", c.IsTLS())
			return c
		},
		Addr: addr.String(),
		ServerConfigCallback: func(ctx ssh.Context) *gossh.ServerConfig {
			return &gossh.ServerConfig{
				BannerCallback: func(conn gossh.ConnMetadata) string {
					return fmt.Sprintf("Olá %s, você está conectado aos servidores SOCXY Cloud (%s). Seu IP: %s",
						conn.User(), conn.ServerVersion(), conn.RemoteAddr())
				},
			}
		},
		Version: "Socxy_v0.1",
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}