package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"github.com/dimfeld/httptreemux"
	"github.com/zenoss/glog"
	"net"
	"net/http"
	"strings"
)

type HookHandler func(http.ResponseWriter, *http.Request, map[string]string, *Hook)

func hookHandler(w http.ResponseWriter, r *http.Request, params map[string]string, hook *Hook) {
	githubEventType := r.Header.Get("X-GitHub-Event")

	if r.ContentLength > 16384 {
		// We should never get a request this large.
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	buffer := bytes.Buffer{}
	buffer.ReadFrom(r.Body)
	r.Body.Close()

	if glog.V(2) {
		niceBuffer := &bytes.Buffer{}
		json.Indent(niceBuffer, buffer.Bytes(), "", "  ")
		glog.Infof("Hook %s received data %s\n",
			r.URL.Path, string(niceBuffer.Bytes()))
	}

	if hook.Secret != "" {
		secret := r.Header.Get("X-Hub-Signature")
		if !strings.HasPrefix(secret, "sha1=") {
			glog.Warningf("Request with no secret for hook %s from %s\n",
				r.URL.Path, r.RemoteAddr)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		hash := hmac.New(sha1.New, []byte(hook.Secret))
		expected := hash.Sum(buffer.Bytes())
		if !hmac.Equal(expected, []byte(secret[5:])) {
			glog.Warningf("Request with bad secret for hook %s from %s\n",
				r.URL.Path, r.RemoteAddr)
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	event, err := NewEvent(buffer.Bytes(), githubEventType)
	if err != nil {
		glog.Errorf("Error parinsg JSON for %s: %s", r.URL.Path, err)
	}
	event["urlparams"] = params
	hook.Execute(event)
}

func handlerWrapper(handler HookHandler, hook *Hook) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		glog.Infoln("Called", r.URL.Path)
		handler(w, r, params, hook)
	}
}

func RunServer(config *Config) {

	var listener net.Listener = nil

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		glog.Fatal("Could not listen on", config.ListenAddress)
	}

	if len(config.AcceptIp) != 0 {
		listenFilter := NewListenFilter(listener, WhiteList)
		for _, a := range config.AcceptIp {
			listenFilter.FilterAddr[a] = true
		}
		listener = listenFilter
	}

	router := httptreemux.New()

	for _, hook := range config.Hook {
		router.POST(hook.Url, handlerWrapper(hookHandler, hook))
	}

	glog.Fatal(http.Serve(listener, router))
}
