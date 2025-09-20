package xutils

import (
	"strconv"
	"strings"
)

func GetFullUrl(domain string, path string, port int, isSecure bool) string {
	if after, ok := strings.CutPrefix(domain, "*."); ok {
		domain = after
	}

	if port != 0 {

		if !isSecure && port != 80 {
			domain = domain + ":" + strconv.Itoa(port)
		} else if isSecure && port != 443 {
			domain = domain + ":" + strconv.Itoa(port)
		}

	}

	if isSecure {
		domain = "https://" + domain
	} else {
		domain = "http://" + domain
	}

	return domain + path
}
