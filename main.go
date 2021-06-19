package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/raydcast/socxy-go/proxy"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Ports []int `yaml:"ports"`

		Certificate struct {
			RSABits  int    `yaml:"rsa_bits"`
			Duration int    `yaml:"duration"`
			SNI      string `yaml:"sni"`
		} `yaml:"certificate"`
	} `yaml:"server"`

	UDPGW struct {
		Host                    string `yaml:"host"`
		Port                    int    `yaml:"port"`
		MaxClients              int    `yaml:"max_clients"`
		MaxConnectionsForClient int    `yaml:"max_connections_for_client"`
	} `yaml:"udpgw"`
}

func main() {
	var (
		configPath string
		config     Config
	)

	flag.StringVar(&configPath, "config", "config.yaml", "Socxy configuration filename")
	flag.Parse()

	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error to open config file: %v", err)
	}

	if err := yaml.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("%v", err)
	}

	cert := proxy.NewCert(config.Server.Certificate.RSABits,
		config.Server.Certificate.Duration, config.Server.Certificate.SNI)

	for _, port := range config.Server.Ports {
		go listen(port, cert)
		log.Printf("Listening server on port: %v", port)
	}

	go udpgw_listen(config.UDPGW.Host, int64(config.UDPGW.Port),
		int64(config.UDPGW.MaxClients), int64(config.UDPGW.MaxConnectionsForClient))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
