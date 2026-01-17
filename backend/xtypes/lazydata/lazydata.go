package lazydata

type LazyData interface {
	AsBytes() ([]byte, error)

	AsMap() (map[string]any, error)
	// AsJSON struct target
	AsJson(target any) error

	// if layzdata is byte type use gjson, if its lua table then get field value
	GetFieldAsInt(path string) int
	GetFieldAsFloat(path string) float64
	GetFieldAsString(path string) string
	GetFieldAsBool(path string) bool
}
