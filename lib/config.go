package lib

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/pelletier/go-toml"
)

type AddressPair struct {
	ListenAddr string
	RemoteAddr string
}
type Config struct {
	// If specified, this channel will be sent a `true` value after all
	// connections to the SSH servers have been established.
	Ready chan bool

	// If true, log verbosely.
	Verbose bool

	connections map[string][]AddressPair
}

func ParseConfig(path string) (Config, error) {
	var config Config
	config.connections = make(map[string][]AddressPair)
	configFile, err := toml.LoadFile(path)
	if err != nil {
		log.Fatalf("parse config: %v", err)
	}
	for _, key := range configFile.Keys() {
		var bastionAddr string
		var addressPair AddressPair

		bastionAddr, ok := configFile.Get(key + ".bastion").(string)
		if !ok {
			return config, errors.New("Top-level key 'bastion' must be a string")
		}
		// Bastion host must be of the form host:port
		if !strings.Contains(bastionAddr, ":") {
			bastionAddr = bastionAddr + ":22"
		}

		if listenAddr, ok := configFile.Get(key + ".listen").(string); ok {
			if !strings.Contains(listenAddr, ":") {
				listenAddr = "127.0.0.1:" + listenAddr
			}
			addressPair.ListenAddr = listenAddr
		} else {
			return config, fmt.Errorf("Section %s: key 'listen' must be a string value", key)
		}

		if remoteAddr, ok := configFile.Get(key + ".forward").(string); ok {
			addressPair.RemoteAddr = remoteAddr
		} else {
			return config, fmt.Errorf("Section %s: key 'forward' must be a string", key)
		}

		config.connections[bastionAddr] = append(config.connections[bastionAddr], addressPair)
	}
	return config, nil
}
