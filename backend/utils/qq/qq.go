package qq

import "github.com/k0kubun/pp"

var (
	Enabled = true
)

func Println(a ...interface{}) (n int, err error) {
	if !Enabled {
		return 0, nil
	}

	return pp.Println(a...)
}
