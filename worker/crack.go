package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"assign2/internal/constants"
	"assign2/internal/messages"

	"golang.org/x/crypto/bcrypt"
)

// Shared state for threads (kept inside crack.go)
var (
	crackCtx    context.Context
	crackCancel context.CancelFunc
	resOnce     sync.Once
	finalRes    *messages.ResultMsg

	totalTested uint64
	crackDone atomic.Uint32
    crackFound atomic.Uint32
)


func crack(job *messages.JobMsg,  numT int) *messages.ResultMsg {
	if numT <= 0 {
		numT = 1
	}

	charsetBytes := []byte(job.Charset)
	base := len(charsetBytes)
	if base == 0 {
		return &messages.ResultMsg{Type: messages.RESULT, Status: "ERROR", Error: "empty charset"}
	}

	crackCtx, crackCancel = context.WithCancel(context.Background())
	defer crackCancel()

	finalRes = &messages.ResultMsg{Type: messages.RESULT, Status: "NOT_FOUND"}



	q := base / numT        // floor
	r := base % numT        // remainder (first r threads get +1)

	var wg sync.WaitGroup
	fmt.Println("[worker] creating cracking threads")
	for tid := 0; tid < numT; tid++ {
		// start = tid*q + min(tid, r)
		startIdx := tid*q
		if tid < r {
			startIdx += tid
		} else {
			startIdx += r
		}

		// len = q + (tid < r ? 1 : 0)
		chunk := q
		if tid < r {
			chunk++
		}

		endIdx := startIdx + chunk

		wg.Go(func() {
			fmt.Printf("[worker] t%d: startIdx=%d endIdx=%d\n",tid, startIdx, endIdx)
			crackThread(job, startIdx, endIdx)
		})
	}

	wg.Wait()

	return finalRes
}


func crackThread(job *messages.JobMsg, startIdx int, endIdx int) {
	charset := []byte(job.Charset)
	base := len(charset)
	ctx := newCryptCtx() 

	for length := 1; length <= constants.MaxPasswordLen; length++ {
		// Stop quickly if another thread found it (or error).
		select {
		case <-crackCtx.Done():
			return
		default:
		}

		verifyCand := func(passCand string) {
			ok, err := verifyCandidate(ctx, job.Alg, passCand, job.FullHash)
				atomic.AddUint64(&totalTested, 1)

				if err != nil {
					setFinalResult(&messages.ResultMsg{Type: messages.RESULT, Status: "ERROR", Error: err.Error()})
					return
				}
				if ok {
					setFinalResult(&messages.ResultMsg{Type: messages.RESULT, Status: "FOUND", Password: passCand})
					return
				}
		}


		suffixLen := length - 1
		suffixSpace := powInt(base, suffixLen)

		for c := startIdx; c < endIdx; c++ {
			select {
			case <-crackCtx.Done():
				return
			default:
			}


			prefix := charset[c]

			if length == 1 {
				verifyCand(string(prefix))
			}

			for s := 0; s < suffixSpace; s++ {
				select {
				case <-crackCtx.Done():
					return
				default:
				}

				suffix := indexToCandidate(s, charset, suffixLen)
				cand := string(prefix) + suffix

				verifyCand(cand)
			}
		}
	}
}

// setFinalResult sets finalRes once and cancels other threads.
func setFinalResult(r *messages.ResultMsg) {
	resOnce.Do(func() {
		finalRes = r
		if r.Status == "FOUND" {
			crackFound.Store(1)
		}
		crackDone.Store(1)
		crackCancel()
	})
}

func powInt(base int, exp int) int {
	out := 1
	for i := 0; i < exp; i++ {
		out *= base
	}
	return out
}


func indexToCandidate(idx int, charset []byte, length int) string {
	base := len(charset)
	out := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		out[i] = charset[idx%base]
		idx /= base
	}
	return string(out)
}

func verifyCandidate(ctx *cryptCtx, alg, candidate, fullHash string) (bool, error) {
	switch alg {
	case "bcrypt":
		err := bcrypt.CompareHashAndPassword([]byte(fullHash), []byte(candidate))
		if err == nil {
			return true, nil
		}
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err

	case "md5", "sha256", "sha512", "yescrypt":
		got, err := cryptHashWithCtx(ctx, candidate, fullHash) // fullHash is the “setting”
		if err != nil {
			return false, err
		}
		return got == fullHash, nil

	default:
		return false, fmt.Errorf("unsupported alg: %s", strings.TrimSpace(alg))
	}
}
