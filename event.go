package main

import (
	"encoding/json"
	"github.com/zenoss/glog"
)

type Event map[string]interface{}

// Information to translate from GitLab to GitHub format.
var eventTranslate = map[string]string{
	"total_commits_count": "size",
	"repository":          "repo",
	"after":               "head",
}

var commitTranslate = map[string]string{
	"id": "sha",
}

func (e Event) Commits() []interface{} {
	generic, ok := e["commits"]
	if !ok {
		if glog.V(1) {
			glog.Infoln("Event had no commits")
		}
		return nil
	}

	interfaceList, ok := generic.([]interface{})
	if !ok {
		glog.Errorf("Commit list had type %T\n", generic)
		return nil
	}

	return interfaceList
}

// Normalize makes GitLab events look like GitHub events.
func (e Event) normalize() {
	// Fortunately there's very little difference, at least in the push format.

	for gitlabName, githubName := range eventTranslate {
		value, ok := e[gitlabName]
		if ok {
			e[githubName] = value
		}
	}

	commits := e.Commits()
	if commits != nil {
		for _, generic := range commits {
			c, ok := generic.(map[string]interface{})
			if !ok {
				glog.Errorf("Commit had type %T", generic)
			}
			for gitlabName, githubName := range commitTranslate {
				value, ok := c[gitlabName]
				if ok {
					c[githubName] = value
				}
			}
		}
	}
}

// Create a new event from the given JSON. If the event type is blank,
// this function will try to figure it out. Generally, GitHub events will
// present a value for this in the HTTP Request, and GitLab events place
// the event type in the JSON.
func NewEvent(jsonData []byte, eventName string) (Event, error) {
	e := Event{}
	err := json.Unmarshal(jsonData, &e)
	if err != nil {
		return nil, err
	}

	if glog.V(4) {
		glog.Infof("Unnormalized event: %v", e)
	}

	if payload, ok := e["object_attributes"].(map[string]interface{}); ok {
		// For GitLab events, export all object_attributes fields into the event scope.
		for key, value := range payload {
			e[key] = value
		}
	}

	e.normalize()

	if glog.V(3) {
		glog.Infof("Event: %v", e)
	}

	if eventName == "" {
		var gitlabType string
		gitlabType, ok := e["object_kind"].(string)
		if !ok {
			// Push events look completely different from the other events and
			// don't have an explicit type.
			gitlabType = "push"
		}
		e["type"] = gitlabType
	} else {
		e["type"] = eventName
	}
	return e, nil
}
