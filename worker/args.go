package main

import (
	"flag"
	"fmt"
)

func parseArgs() (host string, port int, numT int, err error) {
	flag.StringVar(&host, "c", "", "controller host")
	flag.IntVar(&port, "p", 0, "controller port")
	flag.IntVar(&numT, "t", 0, "thread number")
	flag.Parse()

	if host == "" || port <= 0 || port > 65535 || numT <= 0  {
		return "", 0, 0, fmt.Errorf("missing required argument")
	}
	return host, port, numT, nil
}