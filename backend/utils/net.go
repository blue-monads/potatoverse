package xutils

import (
	"fmt"
	"net"
)

func GetLocalIPs() []string {
	ips := make([]string, 0)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips
}

func BuildAddr(host, port string) string {
	if port == "8080" || port == "443" {
		port = ""
	}

	return fmt.Sprintf("http://%s:%s", host, port)
}
