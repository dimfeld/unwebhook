package main

import (
	"bytes"
	"github.com/dimfeld/httptreemux"
	"net"
	"net/http"
)

type HookHandler func(http.ResponseWriter, *http.Request, map[string]string, *Hook)

func hookHandler(w http.ResponseWriter, r *http.Request, params map[string]string, hook *Hook) {
	githubEventType := r.Header().Get("X-GitHub-Event")

	if r.ContentLength > 16384 {
		// We should never get a request this large.
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	buffer := bytes.Buffer{}
	buffer.ReadFrom(r.Body)
	r.Body.Close()

	event := NewEvent(buffer.Bytes(), githubEventType)
	event["urlparams"] = params
	hook.Execute(event)
}

func handlerWrapper(handler HookHandler, hook *Hook) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		if logger != nil {
			logger.Println("Called", r.URL.Path)
		}
		handler(w, r, params, hook)
	}
}

func RunServer(config *Config) {

	var listener net.Listener = nil

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		logger.Fatal("Could not listen on", config.ListenAddress)
	}

	if len(config.AcceptIp) != 0 {
		listenFilter := NewListenFilter(listener, WhiteList, logger)
		for _, a := range config.AcceptIp {
			listenFilter.FilterAddr[a] = true
		}
		listener = listenFilter
	}

	router := httptreemux.New()

	for _, hook := range config.Hook {
		router.POST(hook.Url, handlerWrapper(hookHandler, hook))
	}

	logger.Fatal(http.Serve(listener, router))
}
