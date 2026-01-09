package xutils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
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

// zz-12-serverkey.example.com

func BuildExecHost(currHost string, spaceId int64, hosts []string, serverKey string) string {

	bestHost := findBestHost(hosts, currHost)
	if bestHost == "" {
		return currHost
	}

	if strings.Contains(bestHost, "*") {
		prefix := fmt.Sprintf("zz-%d-%s.", spaceId, serverKey)
		return strings.Replace(bestHost, "*.", prefix, 1)
	}

	return bestHost
}

func findBestHost(hosts []string, currHost string) string {
	bestHost := ""
	for _, host := range hosts {
		if strings.Contains(host, "*") {
			maybe := strings.Replace(host, "*.", "", 1)
			if strings.HasPrefix(currHost, maybe) {
				bestHost = host
				break
			}
		}
	}
	return bestHost
}

var spaceIdPattern = regexp.MustCompile(`zz-(\d+)-`)

func ExtractSpaceId(domain string) int64 {
	qq.Println("@extractDomainSpaceId/1", domain)

	if matches := spaceIdPattern.FindStringSubmatch(domain); matches != nil {
		sid, _ := strconv.ParseInt(matches[1], 10, 64)
		return sid
	}
	return 0
}

/*


const currUrl = new URL("http://localhost:3000/");


findBestHost([
    "*.eeeee.com",
    "aa.localhost",
    "specific.com",
    "*.localhost",
    "*.example.com"],
    currUrl.host
)

*/
