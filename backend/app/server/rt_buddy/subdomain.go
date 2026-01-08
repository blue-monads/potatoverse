package rtbuddy

import (
	"strings"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"golang.org/x/net/publicsuffix"
)

func getSubdomain(host string) (string, error) {

	if before, ok := strings.CutSuffix(host, ".localhost"); ok {
		qq.Println("@getSubdomain/0", before)
		return before, nil
	}

	// 2. Get the Registered Domain (e.g., "example.co.uk")
	mainDomain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", err
	}

	// 3. Remove the main domain from the host to get the subdomain
	subdomain := strings.TrimSuffix(host, mainDomain)
	subdomain = strings.TrimSuffix(subdomain, ".") // Remove trailing dot

	qq.Println("@getSubdomain/1", host, mainDomain, subdomain)

	if strings.Contains(subdomain, ".") {
		parts := strings.Split(subdomain, ".")
		subdomain = parts[0]
	}

	qq.Println("@getSubdomain/2", subdomain)

	return subdomain, nil
}
