package main

import (
	"errors"
	"fmt"
	"os"

	"encoding/json"
)

type Cache struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

type Paths struct {
	Api  string `json:"api"`
	Auth string `json:"auth"`
	User string `json:"user"`
}

type Server struct {
	Address string `json:"address"`
}

type Smtp struct {
	Server   string `json:"server"`
	Identity string `json:"identity"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type System struct {
	Root string `json:"root"`
}

type Config struct {
	Cache  Cache  `json:"cache"`
	Paths  Paths  `json:"paths"`
	Server Server `json:"server"`
	Smtp   Smtp   `json:"smtp"`
	System System `json:"system"`
}

var (
	config = Config{
		Cache: Cache{
			Host: "localhost",
			Port: 11211,
		},
		Paths: Paths{
			Api:  "/api",
			Auth: "https://auth.server.com",
			User: "/user",
		},
		Server: Server{
			Address: ":80",
		},
		Smtp: Smtp{
			Server:   "localhost:25",
			Identity: "sender.server.com",
			From:     "sender@server.com",
			To:       "recipient@server.com",
		},
		System: System{
			Root: "/data",
		},
	}
)

// Validates the configuration
// Will not correct the mistakes
func validateConfig() error {
	if config.Paths.Api[len(config.Paths.Api)-1] == '/' || config.Paths.User[len(config.Paths.User)-1] == '/' {
		return errors.New("url paths may not end in '/'")
	}

	if config.System.Root[len(config.System.Root)-1] == '/' {
		return errors.New("system paths may not end in '/'")
	}

	return nil
}

// Loads the memory with configuration from file
// Should be run synchronously before starting server
func loadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open configuration file: %v", err)
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to decode configuration file: %v", err)
	}

	err = validateConfig()
	if err != nil {
		return fmt.Errorf("invalid configuration file: %v", err)
	}

	return nil
}
