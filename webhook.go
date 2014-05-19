package main

import (
	"github.com/dimfeld/goconfig"
	"github.com/dimfeld/httptreemux"
	"os"
)

type Hook struct {
	Cwd      string
	Commands []string
}

type Hooks struct {
	Hook []Hook
}

type Config struct {
	ListenAddress string
	Port          int
	Hook          []Hook
}

var config *Config

func (c *Config) Merge(other *Config) {

}

func main() {
	config = &Config{
		Port: 80,
	}

	goconfig.Load(config, cfgReader, "WEBHOOK")
}
