package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"assign2/internal/messages"
)

//Worker Main
func main() {
	host, port, numT, err := parseArgs()
	if err != nil {
		usage(err)
		os.Exit(2)
	}


	//connect_to_controller
	fmt.Println("[worker] Connecting...")
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		usage(fmt.Errorf("controller unreachable: %w", err))
		os.Exit(2)
	}
	fmt.Println("[worker] Connected")
	defer conn.Close()

	r := bufio.NewReader(conn)

	// receiving and writingmessages
    inbox := MakeWorkerInbox()
    StartWorkerReceiver(r, inbox)

    sendQ := make(chan any, 64)
    writerDone := make(chan struct{})
    writerErr := make(chan error, 1)

    go func() {
        defer close(writerDone)
        for msg := range sendQ {
            if err := messages.Send(conn, msg); err != nil {
                select { case writerErr <- err: default: }
                return
            }
        }
    }()

	// register_worker
	sendQ <- messages.RegisterMsg{
        Type:   messages.REGISTER,
        Worker: hostnameOr("worker"),
    }

    select {
    case ack := <-inbox.Ack:
        if ack.Type != messages.ACK || ack.Status != "OK" {
            close(sendQ); <-writerDone
            return
        }
    case err := <-inbox.Err:
        _ = err
        close(sendQ); <-writerDone
        return
    case err := <-writerErr:
        _ = err
        close(sendQ); <-writerDone
        return
    }

	// WAIT JOB, validate, start cracking in background
	var job messages.JobMsg
    select {
    case job = <-inbox.Job:
        if err := validateJob(&job); err != nil {
            sendQ <- messages.ResultMsg{Type: messages.RESULT, Status: "ERROR", Error: err.Error()}
            close(sendQ); <-writerDone
            return
        }
    case err := <-inbox.Err:
        _ = err
        close(sendQ); <-writerDone
        return
    case err := <-writerErr:
        _ = err
        close(sendQ); <-writerDone
        return
    }

    
    crackDoneCh := make(chan *messages.ResultMsg, 1)
    go func() {
        startCompute := time.Now()
        res := crack(&job, numT)
        res.WorkerComputeNs = time.Since(startCompute).Nanoseconds()
        crackDoneCh <- res
    }()

    //done := make(chan struct{})
	//waiting.StartDots(done, "[worker] cracking password")
	// MAIN LOOP: heartbeat requests OR final result
	var lastTotal uint64

    for {
        select {
        case hb := <-inbox.HbReq:
            total := atomic.LoadUint64(&totalTested)
            delta := total - lastTotal
            lastTotal = total

            fmt.Println("[HB] responded...")

            // Worker replies with progress since last heartbeat:contentReference[oaicite:1]{index=1}
            sendQ <- messages.HeartbeatRes{
                Type:        messages.HEARTBEAT_RES,
                Seq:         hb.Seq,
                DeltaTested: delta,
                Tested:      total,                     // optional
                Done:        crackDone.Load() != 0,
                Found:       crackFound.Load() != 0,
            }

        case res := <-crackDoneCh:
            sendQ <- res
            close(sendQ)
            <-writerDone
            return

        case err := <-inbox.Err:
            _ = err
            close(sendQ)
            <-writerDone
            return

        case err := <-writerErr:
            _ = err
            close(sendQ)
            <-writerDone
            return
        }
    }
}

// consider deleting
func sendError(conn net.Conn, err error) {
	_ = messages.Send(conn, &messages.ResultMsg{
		Type:   messages.RESULT,
		Status: "ERROR",
		Error:  err.Error(),
	})
}

func usage(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	fmt.Fprintln(os.Stderr, "usage: worker -c <controller_host> -p <port> -t <threads>")
}

func hostnameOr(fallback string) string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return fallback
	}
	return h
}
