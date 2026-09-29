package monitor

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Line struct {
	ts      float32
	n       int
	IP      string
	port    int
	Command string
}

// Compiled once: Monitor is called per connection attempt, not per line.
var monitorLine = regexp.MustCompile(`^\+(\d+\.\d+) \[(\d+) ([\d\.]+|\[[0-9a-f\:]+\]|lua):?(\d+)?\] "(.*?)"`)

const (
	monitorDialTimeout   = 5 * time.Second
	monitorHandshakeTTL  = 5 * time.Second
	monitorReconnectWait = 5 * time.Second
)

// parseMonitorLine parses one MONITOR output line.
// Examples:
//
//	IPv4 +1619454979.381488 [1 172.29.1.2:57676] "brpop"
//	IPv6 +1621757323.274428 [9 [::1]:38824] "DEL" "f00dc225-1975-4590-8a37-9b0ea4ec5acc"
//	LUA  +1621757323.279920 [9 lua] "DEL" "f00dc225-1975-4590-8a37-9b0ea4ec5acc"
func parseMonitorLine(resp string) (Line, bool) {
	l := monitorLine.FindStringSubmatch(resp)
	if len(l) != 6 {
		return Line{}, false
	}
	ts, err := strconv.ParseFloat(l[1], 32)
	if err != nil {
		return Line{}, false
	}
	n, err := strconv.Atoi(l[2])
	if err != nil {
		return Line{}, false
	}
	port, err := strconv.Atoi("0" + l[4])
	if err != nil {
		return Line{}, false
	}
	return Line{
		ts:      float32(ts),
		n:       n,
		IP:      l[3],
		port:    port,
		Command: strings.ToUpper(l[5]),
	}, true
}

func (r *RedisServer) Monitor(ctx context.Context, evt func(bool)) (chan Line, chan error) {
	lines := make(chan Line)
	errors := make(chan error)

	go func() {
		defer close(lines)
		defer close(errors)
		for {
			if r.monitorOnce(ctx, lines, errors, evt) {
				return // context done
			}
			timer := time.NewTimer(monitorReconnectWait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()

	return lines, errors
}

// monitorOnce runs a single MONITOR session. It reports true when the
// context is done and no reconnect should be attempted.
func (r *RedisServer) monitorOnce(ctx context.Context, lines chan Line, errors chan error, evt func(bool)) bool {
	fail := func(err error) bool {
		select {
		case errors <- err:
		case <-ctx.Done():
			return true
		}
		evt(false)
		return ctx.Err() != nil
	}

	dialer := &net.Dialer{Timeout: monitorDialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", r.address)
	if err != nil {
		return fail(err)
	}
	defer conn.Close()

	// Unblock the streaming reader below when the context is cancelled.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()

	if err := conn.SetDeadline(time.Now().Add(monitorHandshakeTTL)); err != nil {
		return fail(err)
	}
	reader := bufio.NewReader(conn)
	if r.password != "" {
		if _, err := fmt.Fprintf(conn, "AUTH %s\n", r.password); err != nil {
			return fail(err)
		}
		resp, err := reader.ReadString('\n')
		if err != nil {
			return fail(err)
		}
		if !strings.HasPrefix(resp, "+OK") {
			return fail(fmt.Errorf("auth failed, bad password"))
		}
	}
	if _, err := fmt.Fprintln(conn, "MONITOR"); err != nil {
		return fail(err)
	}
	// Handshake done: stream without deadline, ctx closes the conn.
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return fail(err)
	}
	for {
		resp, err := reader.ReadString('\n')
		if err != nil {
			if ctx.Err() != nil {
				return true
			}
			return fail(fmt.Errorf("monitor can't read %v", err))
		}
		line, ok := parseMonitorLine(resp)
		if !ok {
			continue
		}
		evt(true)
		select {
		case lines <- line:
		case <-ctx.Done():
			return true
		}
	}
}
