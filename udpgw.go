package main

import (
	"log"
	"net"
	"os/exec"
	"strconv"
)

type UDPGWConfig struct {
	Host                    string `yaml:"host"`
	Port                    int    `yaml:"port"`
	MaxClients              int    `yaml:"max-clients"`
	MaxConnectionsForClient int    `yaml:"max-connections-for-client"`
}

func udpgw_listen(config UDPGWConfig) {
	udpgwCmd := exec.Command("badvpn-udpgw", "--listen-addr",
		net.JoinHostPort(config.Host, strconv.FormatInt(int64(config.Port), 10)),
		"--max-clients", strconv.Itoa(config.MaxClients),
		"--max-connections-for-client", strconv.Itoa(config.MaxConnectionsForClient))

	err := udpgwCmd.Start()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("UDPGW enabled on port: %v", config.Port)
	udpgwCmd.Wait()
}
