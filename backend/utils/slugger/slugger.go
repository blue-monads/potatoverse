package slugger

import (
	"regexp"
	"strings"

	"github.com/thoas/go-funk"
)

var (
	SlugExp        = regexp.MustCompile("^[a-z0-9-]*$")
	ReplaceSlugExp = regexp.MustCompile("[^a-z0-9]+")

	// fixme => copy all reserved words from from autoapi and other
	ColumnReservedSlugs = []string{"page", "limit", "cursor"}
)

func Slugify(s string) string {
	if len(s) > 20 {
		s = s[:10]
	}

	return strings.Trim(ReplaceSlugExp.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func IsValidSlug() bool {
	return false
}

func IsDynGroup(value string) bool {
	return deciSlug(value)
}

func IsDyntable(value string) bool {
	return deciSlug(value)
}

func IsDynColumn(value string) bool {
	if funk.ContainsString(ColumnReservedSlugs, value) {
		return false
	}

	return deciSlug(value)
}

func IsAgent(value string) bool {
	return deciSlug(value)
}

func deciSlug(value string) bool {
	if len(value) > 10 {
		return false
	}

	return SlugExp.Match([]byte(value))
}
