package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Ports []int       `yaml:"ports"`
	UDPGW UDPGWConfig `yaml:"udpgw"`
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

	for _, port := range config.Ports {
		go listen(port)
		log.Printf("Listening server on port: %v", port)
	}

	go udpgw_listen(config.UDPGW)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
