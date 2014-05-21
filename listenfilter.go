package main

import (
	"github.com/dimfeld/glog"
	"net"
)

const (
	WhiteList = iota
	BlackList
)

type ListenFilter struct {
	net.Listener
	// BlackList or WhiteList.
	Behavior   int
	FilterAddr map[string]bool
}

func (f *ListenFilter) Accept() (c net.Conn, err error) {
	for {
		c, err = f.Accept()
		if err != nil {
			return
		}

		addr := c.RemoteAddr().String()
		configured := f.FilterAddr[addr]

		if (configured && f.Behavior == WhiteList) ||
			(!configured && f.Behavior == BlackList) {

			return
		}

		c.Close()

		glog.Infoln("Denied connection from", addr)
	}
}

func NewListenFilter(l net.Listener, behavior int) *ListenFilter {
	return &ListenFilter{
		Listener:   l,
		Behavior:   behavior,
		FilterAddr: make(map[string]bool),
	}
}
