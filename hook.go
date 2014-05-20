package main

import (
	"bytes"
	"text/template"
)

// Hook is defined in webhook.go.

func (hook *Hook) Execute(e Event) {
	allowed := true
	if len(hook.AcceptEvent) != 0 {
		allowed = false
		if eventType, ok := allowedEvent["type"].(string); ok {
			for _, allowedEvent := range hook.AllowEvent {
				if allowedEvent == eventType {
					allowed = true
					break
				}
			}
		}
	}

	if hook.PerCommit {
		commits := e.Commits()
		for _, c := range commits {
			hook.processEvent(c)
		}
	} else {
		hook.processEvent(e)
	}
}

func (hook *Hook) processEvent(e Event) {
	cmd := make([]string, len(hook.template))
	for i, t := range hook.template {
		cmd[i], err = hook.processCommand(e, t)
		if err != nil {
			return err
		}
	}

	hook.runCommand(cmd)
}

func (hook *Hook) processCommand(e Event, t *template.Template) (string, error) {
	cmd := bytes.Buffer{}
	err := t.Execute(cmd, e)
	if err != nil {
		return "", err
	}

	return cmd.Bytes(), nil
}

func (hook *Hook) runCommand(cmd string) {

}

func (hook *Hook) CreateTemplates() error {
	hook.template = make([]*text.Template, len(hook.Commands))
	for i, cmd := range hook.Commands {
		var err error
		hook.template[i], err = template.New("tmpl").Parse(cmd)
		if err != nil {
			hook.template = nil
			return err
		}
	}
}
