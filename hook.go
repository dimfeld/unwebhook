package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dimfeld/glog"
	"os"
	"os/exec"
	"text/template"
	"time"
)

// Hook is defined in webhook.go.

var templateFuncs = template.FuncMap{
	"json": func(obj interface{}) string {
		result, err := json.Marshal(obj)
		if err != nil {
			return "<< " + err.Error() + " >>"
		}
		return string(result)
	},
}

// CreateTemplates parses the Commands array into templates.
func (hook *Hook) CreateTemplates() error {
	var err error
	hook.template = make([][]*template.Template, len(hook.Commands))
	for i, cmdList := range hook.Commands {
		hook.template[i] = make([]*template.Template, len(cmdList))

		for j, cmd := range cmdList {

			hook.template[i][j], err = template.New("tmpl").Funcs(templateFuncs).Parse(cmd)
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
				glog.Warningf("Hook %s got disallowed event type %s\n", hook.Url, eventType)
				return
			}
		}

	}

	if hook.PerCommit {
		commits := e.Commits()
		if commits != nil {
			for _, generic := range commits {
				c, ok := generic.(map[string]interface{})
				if !ok {
					glog.Errorf("Commit had type %T", generic)
					continue
				}

				// Set the current commit to pass to the hook.
				e["commit"] = c

				err := hook.processEvent(e)
				if err != nil {
					glog.Errorf("Error processing %s: %s\n", hook.Url, err)
					if glog.V(1) {
						glog.Info(e)
					}
				}
			}
		}
	} else {
		err := hook.processEvent(e)
		if err != nil {
			glog.Errorf("Error processing %s: %s\n", hook.Url, err)
			if glog.V(1) {
				glog.Info(e)
			}
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
	glog.Infoln("Running", args)
	cmd := exec.Command(args[0], args[1:]...)
	if len(hook.Env) != 0 {
		cmd.Env = hook.Env
	}
	cmd.Dir = hook.Dir
	// TODO Make these redirectable
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
