package waiting

import (
	"fmt"
	"strings"
	"time"
)

func StartSpinner(done <-chan struct{}, action string) {
    go func() {
		
		spinner := []rune{'|', '/', '-', '\\'}
        for i := 0;;i++ {
            select {
            case <-done:
                fmt.Printf("\n%s done", action)
                return
            default:
                fmt.Printf("\r%s %c", action, spinner[i])
				if i >= 3 {
					i = 0
				}
    			time.Sleep(200 * time.Millisecond)
            }
        }
    }()
}

func StartDots(done <-chan struct{}, action string) {
    go func() {
		
        for i := 0;;i++ {
            select {
            case <-done:
                fmt.Printf("\n%s done", action)
                return
            default:
                fmt.Printf("\r%s %s      ", action, strings.Repeat(".", i))
				if i >= 4 {
					i = 0
				}
    			time.Sleep(500 * time.Millisecond)
            }
        }
    }()
}
