package executor

// NewShell is a wrapper constructor for NewLocal or NewRemote executor depending on ip provided.
func NewShell(ip string) (Executor, error) {
	if ip == "127.0.0.1" || ip == "localhost" {
		return NewLocal(), nil
	}
	return NewRemoteFromIP(ip)
}
