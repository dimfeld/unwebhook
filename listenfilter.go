package main

import (
	"log"
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
	Logger     *log.Logger
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

		if f.Logger != nil {
			f.Logger.Println("Denied connection from", addr)
		}
	}
}

func NewListenFilter(l net.Listener, behavior int, logger *log.Logger) *ListenFilter {
	return &ListenFilter{
		Listener:   l,
		Behavior:   behavior,
		Logger:     logger,
		FilterAddr: make(map[string]bool),
	}
}
