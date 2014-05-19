package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/dimfeld/goconfig"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Hook struct {
	// Cwd The
	Cwd string
	// Accept notifications from the listed servers.
	// Valid values are "gitlab" and "github". By default,
	// all servers are allowed.
	AcceptServer []string
	// Accept notifications for the following events.
	// Valid values are "push", "newissue"
	AcceptEvent []string
	Commands    []string
}

type Hooks struct {
	Hook []Hook
}

type Config struct {
	ListenAddress string
	Port          int

	LogFile   string
	LogPrefix string

	HookPath []string

	Hook []Hook
}

var logger log.Logger

func (c *Config) MergeHooks(other Hooks) {
	c.Hook = append(c.Hook, other.Hook...)
}

func (c *Config) AddHookFile(filepath string) {
	h := Hooks{}

	f, err := os.Open(filepath)
	if err != nil {
		logger.Logf("Error loading %s: %s", filepath, err)
		return
	}
	defer f.Close()

	_, err = toml.DecodeReader(f, h)
	if err != nil {
		logger.Logf("Error loading %s: %s", filepath, err)
		return
	}

	c.MergeHooks(h)
}

func (c *Config) AddHookPath(p string) {
	info, err := os.Stat(p)
	if err != nil {
		logger.Logf("Error loading %s: %s", p, err)
		return
	}

	if info.IsDir() {
		filepath.Walk(p,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					logger.Logf("Error loading %s, %s", path, err)
					return err
				}
				if info.IsDir() {
					return nil
				}

				c.AddHookFile(path)
			})
	} else {
		c.AddHookFile(p)
	}
}

func main() {
	config = &Config{
		Port: 80,
	}

	mainConfigPath := os.Getenv("WEBHOOK_CONFFILE")
	hooksStartIndex := 1
	if mainConfigPath == "" {
		if len(os.Args) > 1 {
			mainConfigPath := os.Args[1]
		} else {
			mainConfigPath := os.Args[0] + ".conf"
		}
		hooksStartIndex = 2
	}

	if mainConfigPath == "-" {
		goconfig.Load(config, os.Stdin, "WEBHOOK")
	} else {
		f, err := os.Open(mainConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open config file %s: %s".
				mainConfigPath, err)
			os.Exit(1)
		}
		err = goconfig.Load(config, f, "WEBHOOK")
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file %s: %s",
				mainConfigPath, err)
			os.Exit(1)
		}
	}

	for h := range config.HookPath {
		config.AddHookPath(h)
	}

	if len(os.Args) > hooksStartIndex {
		for arg := range os.Args[hooksStartIndex:] {
			config.AddHookPath(arg)
		}
	}

	RunServer(config)
}
