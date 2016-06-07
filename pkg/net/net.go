package net

// IsAddrLocal returns true when given address points to local machine.
func IsAddrLocal(addr string) bool {
	return addr == "127.0.0.1" || addr == "localhost"
}
