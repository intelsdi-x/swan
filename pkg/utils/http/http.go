package http

import (
	"net"
	"time"
)

const retries = 5

// IsListeningFunction is a function type for checking if http endpoint is responding.
type IsListeningFunction func(address string, timeout time.Duration) bool

// IsListeningMockedSuccess is a mocked IsListeningFunction returning always true.
func IsListeningMockedSuccess(address string, timeout time.Duration) bool {
	return true
}

// IsListeningMockedFailure is a mocked IsListeningFunction returning always false.
func IsListeningMockedFailure(address string, timeout time.Duration) bool {
	return false
}

// IsListening tries to connect to given address http address in a form of `http://ip:port`.
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
