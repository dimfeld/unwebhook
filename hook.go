package main

import (
	"bytes"
	"text/template"
)

// Hook is defined in webhook.go.

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
			}
		}
	} else {
		err := hook.processEvent(e)
		if err != nil {
			logger.Printf("Error processing %s: %s", hook.Url, err)
		}
	}
}

func (hook *Hook) processEvent(e Event) error {
	cmds := make([]string, len(hook.template))
	for i, t := range hook.template {
		var err error
		cmds[i], err = hook.processCommand(e, t)
		if err != nil {
			return err
		}
	}

	for _, cmd := range cmds {
		err := hook.runCommand(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (hook *Hook) processCommand(e Event, t *template.Template) (string, error) {
	cmd := &bytes.Buffer{}
	err := t.Execute(cmd, e)
	if err != nil {
		return "", err
	}

	return string(cmd.Bytes()), nil
}

func (hook *Hook) runCommand(cmd string) error {
	return nil
}

func (hook *Hook) CreateTemplates() error {
	hook.template = make([]*template.Template, len(hook.Commands))
	for i, cmd := range hook.Commands {
		var err error
		hook.template[i], err = template.New("tmpl").Parse(cmd)
		if err != nil {
			hook.template = nil
			return err
		}
	}
	return nil
}
