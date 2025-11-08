package xtypes

type LazyData interface {
	AsMap() (map[string]any, error)
	// AsJSON struct target
	AsJson(target any) error
}
