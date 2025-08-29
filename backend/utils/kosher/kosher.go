package kosher

import (
	"github.com/siddontang/go/hack"
)

var USAFE bool = true

func Str(data []byte) string {
	if !USAFE {
		return string(data)
	}
	return hack.String(data)
}

func Byte(data string) []byte {
	if !USAFE {
		return []byte(data)
	}
	return hack.Slice(data)
}
