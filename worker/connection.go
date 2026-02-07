package main

import (
	"bufio"
	"encoding/json"
	"fmt"

	"assign2/internal/messages"
)

// WorkerInbox channels let your worker react without peeking the socket directly.
type WorkerInbox struct {
	Ack   chan messages.AckMsg
	Job   chan messages.JobMsg
	HbReq chan messages.HeartbeatReq
	Err   chan error
}

// MakeWorkerInbox creates buffered channels to avoid deadlocks.
func MakeWorkerInbox() *WorkerInbox {
	return &WorkerInbox{
		Ack:   make(chan messages.AckMsg, 8),
		Job:   make(chan messages.JobMsg, 2),
		HbReq: make(chan messages.HeartbeatReq, 32),
		Err:   make(chan error, 1),
	}
}

// StartWorkerReceiver starts ONE thread that continuously reads messages from conn,
func StartWorkerReceiver(r *bufio.Reader, inbox *WorkerInbox) {
	go func() {
		for {
			raw, err := messages.RecvRawLine(r)
			if err != nil {
				inbox.Err <- fmt.Errorf("recv failed: %w", err)
				return
			}

			t, err := messages.PeekType(raw)
			if err != nil {
				inbox.Err <- fmt.Errorf("peek type failed: %w", err)
				return
			}

			switch t {

			case messages.ACK:
				var m messages.AckMsg
				if err := json.Unmarshal(raw, &m); err != nil {
					inbox.Err <- fmt.Errorf("bad ACK: %w", err)
					return
				}
				inbox.Ack <- m

			case messages.JOB:
				var m messages.JobMsg
				if err := json.Unmarshal(raw, &m); err != nil {
					inbox.Err <- fmt.Errorf("bad JOB: %w", err)
					return
				}
				inbox.Job <- m

			case messages.HEARTBEAT_REQ:
				var m messages.HeartbeatReq
				if err := json.Unmarshal(raw, &m); err != nil {
					inbox.Err <- fmt.Errorf("bad HEARTBEAT_REQ: %w", err)
					return
				}
				fmt.Println("[HB] request received...")
				inbox.HbReq <- m

			default:
				inbox.Err <- fmt.Errorf("unknown message type: %s", t)
				return
			}
		}
	}()
}
