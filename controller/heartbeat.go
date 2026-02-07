package main

import (
	"net"
	"time"

	"assign2/internal/messages"
)

// StartHeartbeatSender sends HEARTBEAT_REQ every hbSec seconds until stop is closed.
// It only WRITES to the connection.
func StartHeartbeatSender(conn net.Conn, hbSec int, stop <-chan struct{}) {
	if hbSec <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(hbSec) * time.Second)
	defer ticker.Stop()

	var seq uint64 = 0

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			seq++
			req := messages.HeartbeatReq{
				Type: messages.HEARTBEAT_REQ,
				Seq:  seq,
			}
			// If send fails, the receiver will usually hit an error shortly after anyway.
			_ = messages.Send(conn, req)
		}
	}
}
