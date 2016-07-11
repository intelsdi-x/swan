package netutil

import (
	"net"
	"time"
)

const retries = 5

// IsListeningFunction is a function type for checking if http endpoint is responding.
type IsListeningFunction func(address string, timeout time.Duration) bool

// IsListening tries to establish TCP connection to given address in a form of `ip:port`.
// It returns true when it was able to connect to given endpoint within timeout time.
func IsListening(address string, timeout time.Duration) bool {
	sleepTime := time.Duration(
		timeout.Nanoseconds() / int64(retries))
	connected := false
	for i := 0; i < retries; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			time.Sleep(sleepTime)
			continue
		}
		defer conn.Close()
		connected = true
	}

	return connected
}
