package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"
	"time"
)

// Hook is defined in webhook.go.

// CreateTemplates parses the Commands array into templates.
func (hook *Hook) CreateTemplates() error {
	var err error
	hook.template = make([][]*template.Template, len(hook.Commands))
	for i, cmdList := range hook.Commands {
		hook.template[i] = make([]*template.Template, len(cmdList))

		for j, cmd := range cmdList {

			hook.template[i][j], err = template.New("tmpl").Parse(cmd)
			if err != nil {
				hook.template = nil
				return err
			}
		}
	}
	return nil
}

// Execute a hook with the given event.
func (hook *Hook) Execute(e Event) {
	if len(hook.AllowEvent) != 0 {
		if eventType, ok := e["type"].(string); ok {
			allowed := false
			for _, allowedEvent := range hook.AllowEvent {
				if allowedEvent == eventType {
					allowed = true
					break
				}
			}

			if !allowed {
				logger.Println("Hook %s got disallowed event type %s", hook.Url, eventType)
				return
			}
		}

	}

	if hook.PerCommit {
		commits := e.Commits()
		for _, c := range commits {
			err := hook.processEvent(c)
			if err != nil {
				logger.Printf("Error processing %s: %s", hook.Url, err)
				debug(e)
			}
		}
	} else {
		err := hook.processEvent(e)
		if err != nil {
			logger.Printf("Error processing %s: %s", hook.Url, err)
			debug(e)
		}
	}
}

func (hook *Hook) processEvent(e Event) error {
	cmds := make([][]string, len(hook.template))
	for i, t := range hook.template {
		var err error
		cmds[i], err = hook.processCommand(e, t)
		if err != nil {
			return err
		}

		execPath, err := exec.LookPath(cmds[i][0])
		if err != nil {
			return fmt.Errorf("Executable %s %s", cmds[i][0], err)
		}
		cmds[i][0] = execPath
	}

	for _, cmd := range cmds {
		err := hook.runCommand(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (hook *Hook) processCommand(e Event, templateList []*template.Template) ([]string, error) {
	cmdList := make([]string, len(templateList))

	for i, t := range templateList {
		cmd := &bytes.Buffer{}

		err := t.Execute(cmd, e)
		if err != nil {
			return nil, err
		}

		cmdList[i] = string(cmd.Bytes())
	}

	return cmdList, nil
}

func (hook *Hook) runCommand(args []string) error {
	debug("Running", args)
	cmd := exec.Command(args[0], args[1:]...)
	if len(hook.Env) != 0 {
		cmd.Env = hook.Env
	}
	cmd.Dir = hook.Dir

	done := make(chan int, 1)

	cmd.Start()
	go func() {
		cmd.Wait()
		done <- 1
	}()

	timer := time.NewTimer(time.Duration(hook.Timeout) * time.Second)

	select {
	case <-done:
		timer.Stop()
		return nil

	case <-timer.C:
		cmd.Process.Kill()
		return fmt.Errorf("Command %v timed out", args)
	}

}
