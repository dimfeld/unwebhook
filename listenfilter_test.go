package main

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

type FakeConn struct {
	HostPort string
	Closed   bool
}

type TestCase struct {
	HostPort string
	Match    bool
}

func TestListenFilter(t *testing.T) {
	behaviors := []int{WhiteList, BlackList}

	fl := &FakeListener{}

	addrs := []string{
		"192.168.1.2",
		"10.1.0.0/16",
		"10.2.1.1/16",
		"200.3.4.5/8",
		"1.2.3.4/32",
		"2000::",
		"1000::/64",
		"4000::2468/128",
		"4000::abab",
	}

	testcases := []TestCase{
		TestCase{"192.168.1.2:1243", true},
		TestCase{"192.168.1.1:734", false},
		TestCase{"192.168.1.0:23", false},
		TestCase{"10.1.78.45:2345", true},
		TestCase{"10.1.0.0:345", true},
		TestCase{"10.1.255.255:345", true},
		TestCase{"10.2.0.0:1", true},
		TestCase{"10.2.89.89:65535", true},
		TestCase{"10.3.0.0:5900", false},
		TestCase{"200.3.0.0:6000", true},
		TestCase{"200.0.0.0:23", true},
		TestCase{"200.3.4.5:345", true},
		TestCase{"201.0.0.0:12", false},
		TestCase{"201.255.255.255:345", false},
		TestCase{"[2000::]:894", true},
		TestCase{"[2000::1]:7682", false},
		TestCase{"[2001::1]:7682", false},
		TestCase{"[1000::]:7682", true},
		TestCase{"[1000::1]:7682", true},
		TestCase{"[1000::1000:0000:0000:0000]:7682", true},
		TestCase{"[1000::1000:0000:0000:0001]:7682", true},
		TestCase{"[1000:ffff:ffff:ffff::1]:7682", false},
		TestCase{"[1000:0000:0000:0001::]:7682", false},
		TestCase{"[1001::]:7682", false},
		TestCase{"[4000::2468]:7682", true},
		TestCase{"[4000:8000::2468]:345", false},
		TestCase{"[4000::abab]:345", true},
		TestCase{"[4000::abab]:345", true},
		TestCase{"[::]:456", false},
	}

	f := NewListenFilter(fl, WhiteList)

	for _, addr := range addrs {
		err := f.AddString(addr)
		if err != nil {
			t.Errorf("Failed to add address %s: %s", addr, err)
		}
	}

	for _, behavior := range behaviors {
		f.Behavior = behavior

		matchIsAllowed := behavior == WhiteList

		behaviorStr := "whitelist"
		if behavior == BlackList {
			behaviorStr = "blacklist"
		}

		for _, testcase := range testcases {
			expectAllowed := matchIsAllowed == testcase.Match
			fl.ConnAddr = testcase.HostPort
			fl.AcceptedOnce = false

			c, err := f.Accept()
			if err != nil {
				if err.Error() == "Fake Accept" {
					if !expectAllowed {
						continue
					} else {
						t.Errorf("Address %s was incorrectly accepted by %s", testcase.HostPort, behaviorStr)
					}
				} else {
					t.Errorf("Error trying address %s: %s", testcase.HostPort, err)
				}
			}

			if c == nil && expectAllowed {
				t.Errorf("Address %s was incorrectly denied by %s", testcase.HostPort, behaviorStr)
			}
		}
	}
}

func tryFilterAddError(t *testing.T, l *ListenFilter, prefix string) {
	err := l.AddString(prefix)
	if err == nil {
		t.Error("Expected failure when adding", prefix)
	}
}

// TestListenFilterAddError ensures that errors are returned when adding
// invalid IP prefixes.
func TestListenFilterAddError(t *testing.T) {
	l := NewListenFilter(nil, WhiteList)
	tryFilterAddError(t, l, "192.168.1.1.1")
	tryFilterAddError(t, l, "192.168.1.1/24/56")
	tryFilterAddError(t, l, "192.168.1.1/33")
	tryFilterAddError(t, l, "2000::5678::")
	tryFilterAddError(t, l, "2000::5678::/64")
	tryFilterAddError(t, l, "2000::5678/129")
	tryFilterAddError(t, l, "2000::5678//128")
}

func BenchWithOneSubnet(b *testing.B) {
	addr := &FakeListener{"192.168.1.65", false}
	f := NewListenFilter(addr, WhiteList)

	f.AddString("192.168.1.0/24")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Accept()
	}
}

func BenchWithFiftySubnets(b *testing.B) {
	addr := &FakeListener{"192.168.49.65", false}
	f := NewListenFilter(addr, WhiteList)

	for i := 0; i < 50; i++ {
		prefix := fmt.Sprintf("192.168.%d.0/24", i)
		err := f.AddString(prefix)
		if err != nil {
			b.Errorf("Failed adding prefix %s: %s", prefix, err)
			b.FailNow()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr.AcceptedOnce = false
		f.Accept()
	}
}

type SimpleAddr string

func (s SimpleAddr) String() string {
	return string(s)
}

func (s SimpleAddr) Network() string {
	return "tcp"
}

// FakeListener returns a fake connection with its string as the RemoteAddr.
type FakeListener struct {
	ConnAddr     string
	AcceptedOnce bool
}

func (f *FakeListener) Accept() (c net.Conn, err error) {
	if f.AcceptedOnce {
		return nil, errors.New("Fake Accept")
	} else {
		f.AcceptedOnce = true
		return &FakeConn{HostPort: f.ConnAddr}, nil
	}
}

func (f *FakeListener) Close() error {
	return nil
}

func (f *FakeListener) Addr() net.Addr {
	return nil
}

// Dummy implementation of the Conn interface. For this test's purposes,
// almost nothing needs a real implementation.
func (f *FakeConn) Close() error {
	f.Closed = true
	return nil
}

func (f *FakeConn) RemoteAddr() net.Addr {

	return SimpleAddr(f.HostPort)
}

func (f *FakeConn) Read(b []byte) (int, error)         { return 0, nil }
func (f *FakeConn) Write(b []byte) (int, error)        { return len(b), nil }
func (f *FakeConn) LocalAddr() net.Addr                { return nil }
func (f *FakeConn) SetDeadline(t time.Time) error      { return nil }
func (f *FakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (f *FakeConn) SetWriteDeadline(t time.Time) error { return nil }
