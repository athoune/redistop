package monitor

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Line struct {
	ts      float32
	n       int
	IP      string
	port    int
	Command string
}

func Monitor(ctx context.Context, address string, password string) (chan Line, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintln(conn, "PING")
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(resp, "-NOAUTH") {
		fmt.Fprintf(conn, "AUTH %s\n", password)
		resp, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(resp, "+OK") {
			return nil, fmt.Errorf("AUTH not ok : %s", resp)
		}
	}
	_, err = fmt.Fprintln(conn, "MONITOR")
	if err != nil {
		return nil, err
	}
	lines := make(chan Line)
	// +1619454979.381488 [1 172.29.1.2:57676] "brpop"
	line, err := regexp.Compile(`^\+(\d+\.\d+) \[(\d+) ([\d.]+):(\d+)] "(.*?)"`)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			resp, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				break
			}
			l := line.FindStringSubmatch(resp)
			if len(l) != 6 {
				continue
			}
			ts, err := strconv.ParseFloat(l[1], 32)
			if err != nil {
				fmt.Println(err)
				break
			}
			n, err := strconv.Atoi(l[2])
			if err != nil {
				fmt.Println(err)
				break
			}
			port, err := strconv.Atoi(l[4])
			if err != nil {
				fmt.Println(err)
				break
			}
			lines <- Line{
				ts:      float32(ts),
				n:       n,
				IP:      l[3],
				port:    port,
				Command: strings.ToUpper(l[5]),
			}
		}
	}()

	return lines, nil
}
