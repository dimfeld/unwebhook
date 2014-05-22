package main

import (
	"errors"
	"github.com/dimfeld/glog"
	"net"
	"strings"
)

const (
	WhiteList = iota
	BlackList
)

type ListenFilter struct {
	net.Listener
	// BlackList or WhiteList.
	Behavior   int
	FilterNet  []*net.IPNet
	FilterAddr []net.IP
}

func (f *ListenFilter) AddString(s string) error {
	if strings.Contains(s, "/") {
		_, net, err := net.ParseCIDR(s)
		if err != nil {
			return err
		}
		f.FilterNet = append(f.FilterNet, net)
	} else {
		addr := net.ParseIP(s)
		if addr == nil {
			return errors.New("Invalid address")
		}
		f.FilterAddr = append(f.FilterAddr, addr)
	}

	return nil
}

func (f *ListenFilter) Accept() (c net.Conn, err error) {
	for {
		c, err = f.Listener.Accept()
		if err != nil {
			return
		}

		var addrStr string
		addrStr, _, err = net.SplitHostPort(c.RemoteAddr().String())
		addr := net.ParseIP(addrStr)

		// A trie would be better here. But for this program there will rarely
		// be more than one or two entries so it doesn't really matter.

		found := false
		for _, net := range f.FilterNet {
			if net.Contains(addr) {
				found = true
				break
			}
		}

		if !found {
			for _, filterAddr := range f.FilterAddr {
				if filterAddr.Equal(addr) {
					found = true
					break
				}
			}
		}

		if (found && f.Behavior == WhiteList) ||
			(!found && f.Behavior == BlackList) {

			// Connection allowed.
			return
		}

		// Connection denied. Just close it silently.
		c.Close()
		glog.Warningln("Denied connection from", addr)
	}
}

func NewListenFilter(l net.Listener, behavior int) *ListenFilter {
	return &ListenFilter{
		Listener:   l,
		Behavior:   behavior,
		FilterAddr: make([]net.IP, 0),
		FilterNet:  make([]*net.IPNet, 0),
	}
}
