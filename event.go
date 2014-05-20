package main

import (
	"encoding/json"
)

type Event map[string]interface{}

type CommitList []Event

// Information to translate from GitLab to GitHub format.
var eventTranslate = map[string]string{
	"total_commits_count": "size",
	"repository":          "repo",
	"after":               "head",
}

var commitTranslate = map[string]string{
	"id": "sha",
}

func (e Event) Commits() CommitList {
	generic, ok := e["commits"]
	if !ok {
		return nil
	}

	c, ok := generic.(CommitList)
	if !ok {
		return nil
	}

	return c
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
	for _, c := range commits {
		for gitlabName, githubName := range commitTranslate {
			value, ok := c[gitlabName]
			if ok {
				c[githubName] = value
			}
		}
	}

}

// Create a new event from the given JSON. If the event type is blank,
// this function will try to figure it out. Generally, GitHub events will
// present a value for this in the HTTP Request, and GitLab events place
// the event type in the JSON.
func NewEvent(jsonData []byte, eventName string) Event {
	e := Event{}
	json.Unmarshal(jsonData, e)

	if payload, ok := e["payload"].(Event); ok {
		// For GitHub events, export all payload fields into the event scope.
		for key, value := range payload {
			e[key] = value
		}
	}

	e.normalize()

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
	return e
}
