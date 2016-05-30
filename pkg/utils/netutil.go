package utils

func IsLocalAddress(host string) bool {
	return host == "127.0.0.1" || host == "localhost"
}
