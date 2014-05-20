package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/dimfeld/goconfig"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"text/template"
)

type Hook struct {
	// URL at which this hook should be available.
	Url string
	// Dir is the working directory from which the command should be run.
	// If blank, the current working directory is used.
	Dir string
	// Env is a list of environment variables to set. If empty, the current
	// environment is used. Each item takes the form "key=value"
	Env []string

	// If PerCommit is true, call the hook once for each commit in the message.
	// Otherwise it is just called once per message.
	PerCommit bool

	// If empty, all events are accepted.
	AllowEvent []string

	// Commands to run.
	Commands [][]string

	// Override the default timeout.
	Timeout int

	template [][]*template.Template
}

type Hooks struct {
	Hook []*Hook
}

type Config struct {
	ListenAddress string

	DebugMode bool

	// The maximum amount of time to wait for a command to finish.
	// Default is 5 seconds.
	CommandTimeout int

	// Accept connections from only the given IP addresses.
	AcceptIp []string

	LogFile   string
	LogPrefix string

	// Paths to search for hook files
	HookPath []string

	Hook []*Hook
}

var (
	logger    *log.Logger
	debugMode bool
)

func debugf(format string, args ...interface{}) {
	if debugMode {
		logger.Printf(format, args...)
	}
}

func debug(args ...interface{}) {
	if debugMode {
		logger.Println(args...)
	}
}

func (c *Config) MergeHooks(other Hooks) {
	c.Hook = append(c.Hook, other.Hook...)
}

func (c *Config) AddHookFile(file string) {
	h := Hooks{}

	f := os.Stdin

	if file == "-" {
		// Change file here so that any error messages will look better.
		file = "stdin"
	} else {
		f, err := os.Open(file)
		if err != nil {
			logger.Printf("Error loading %s: %s", file, err)
			return
		}
		defer f.Close()
	}

	_, err := toml.DecodeReader(f, h)
	if err != nil {
		logger.Printf("Error loading %s: %s", file, err)
		return
	}

	c.MergeHooks(h)
}

func (c *Config) AddHookPath(p string) {
	info, err := os.Stat(p)
	if err != nil {
		logger.Printf("Error loading %s: %s", p, err)
		return
	}

	if info.IsDir() {
		filepath.Walk(p,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					logger.Printf("Error loading %s, %s", path, err)
					return err
				}
				if info.IsDir() {
					return nil
				}

				c.AddHookFile(path)
				return nil
			})
	} else {
		c.AddHookFile(p)
	}
}

func catchSIGINT(f func(), quit bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			logger.Println("SIGINT received...")
			f()
			if quit {
				os.Exit(1)
			}
		}
	}()
}

func main() {
	config := &Config{
		ListenAddress:  ":80",
		CommandTimeout: 5,
	}

	mainConfigPath := os.Getenv("WEBHOOK_CONFFILE")
	hooksStartIndex := 1
	if mainConfigPath == "" {
		if len(os.Args) > 1 {
			mainConfigPath = os.Args[1]
		} else {
			mainConfigPath = os.Args[0] + ".conf"
		}
		hooksStartIndex = 2
	}

	if mainConfigPath == "-" {
		goconfig.Load(config, os.Stdin, "WEBHOOK")
	} else {
		f, err := os.Open(mainConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open config file %s: %s",
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

	logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open log file %s\n", config.LogFile)
		os.Exit(1)
	}

	logger = log.New(logFile, config.LogPrefix, log.LstdFlags)

	debugMode = config.DebugMode

	for _, h := range config.HookPath {
		config.AddHookPath(h)
	}

	if len(os.Args) > hooksStartIndex {
		for _, arg := range os.Args[hooksStartIndex:] {
			config.AddHookPath(arg)
		}
	}

	closer := func() {
		logFile.Close()
	}
	catchSIGINT(closer, true)
	defer closer()

	failed := false
	for _, h := range config.Hook {
		if h.Timeout == 0 {
			h.Timeout = config.CommandTimeout
		}

		err := h.CreateTemplates()
		if err != nil {
			logger.Printf("Failed parsing template %s: %s", h.Url, err)
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}

	RunServer(config)
}
