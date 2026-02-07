package main

import (
	"flag"
	"fmt"
)

func parseArgs() (shadowPath, username string, port int, hbSec int, err error) {
	flag.StringVar(&shadowPath, "f", "", "path to shadow file")
	flag.StringVar(&username, "u", "", "username")
	flag.IntVar(&port, "p", 0, "port")
	flag.IntVar(&hbSec, "b", 0, "heartbeat seconds")
	flag.Parse()

	if shadowPath == "" || username == "" || port <= 0 || port > 65535 || hbSec <=0 {
		return "", "", 0, 0, fmt.Errorf("missing required argument")
	}
	return shadowPath, username, port, hbSec, nil
}
