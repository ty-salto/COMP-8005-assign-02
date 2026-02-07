// // heartbeat.go
package main

// import (
// 	"bufio"
// 	"encoding/json"
// 	"net"
// 	"sync"
// 	"sync/atomic"
// 	"time"

// 	"assign2/internal/messages"
// )

// // Responds ONLY when a HEARTBEAT_REQ arrives.
// // Keeps the connection open; does NOT send heartbeats proactively.
// func heartbeatResponder(
// 	conn net.Conn,
// 	r *bufio.Reader,
// 	writeMu *sync.Mutex,
// 	stop <-chan struct{},
// 	totalTested *uint64,      // atomic counter (your totalTested in crack.go)
// 	crackDone *atomic.Uint32, // 0/1
// 	crackFound *atomic.Uint32, // 0/1
// ) {
// 	var lastTotal uint64

// 	for {
// 		// Let us stop even if controller stops sending heartbeats.
// 		_ = conn.SetReadDeadline(time.Now().Add(250 * time.Millisecond))

// 		raw, err := messages.RecvRawLine(r)
// 		if err != nil {
// 			// Timeout is expected; use it to check stop and keep waiting.
// 			if ne, ok := err.(net.Error); ok && ne.Timeout() {
// 				select {
// 				case <-stop:
// 					return
// 				default:
// 					continue
// 				}
// 			}
// 			// Other errors = connection closed/broken.
// 			return
// 		}

// 		typ, err := messages.PeekType(raw)
// 		if err != nil {
// 			continue
// 		}

// 		if typ != messages.HEARTBEAT_REQ {
// 			// During cracking, you can ignore anything that isn't heartbeat.
// 			// (Controller shouldn't send other types at this time anyway.)
// 			continue
// 		}

// 		var req messages.HeartbeatReq
// 		if err := json.Unmarshal(raw, &req); err != nil {
// 			continue
// 		}

// 		total := atomic.LoadUint64(totalTested)
// 		delta := total - lastTotal
// 		lastTotal = total

// 		res := messages.HeartbeatRes{
// 			Type:        messages.HEARTBEAT_RES,
// 			Seq:         req.Seq,
// 			DeltaTested: delta, // <-- add this field in messages.go
// 			Tested:      total, // optional but useful
// 			Done:        crackDone.Load() != 0,
// 			Found:       crackFound.Load() != 0,
// 		}

// 		writeMu.Lock()
// 		_ = messages.Send(conn, res)
// 		writeMu.Unlock()

// 		// keep looping and responding to every request until stop is closed
// 	}
// }
