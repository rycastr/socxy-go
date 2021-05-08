package main

import (
	"log"
	"net"
	"os/exec"
	"strconv"
)

func udpgw_listen(host string, port, maxClients, maxConnectionsForClient int64) {
	udpgwCmd := exec.Command("./tools/badvpn-udpgw", "--listen-addr",
		net.JoinHostPort(host, strconv.FormatInt(port, 10)),
		"--max-clients", strconv.FormatInt(maxClients, 10),
		"--max-connections-for-client", strconv.FormatInt(maxConnectionsForClient, 10))

	err := udpgwCmd.Start()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("UDPGW enabled on port: %v", port)
	udpgwCmd.Wait()
}
