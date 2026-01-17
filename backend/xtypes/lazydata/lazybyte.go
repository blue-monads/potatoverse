package lazydata

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

type LazyDataBytes []byte

func (l LazyDataBytes) AsBytes() ([]byte, error) {
	return l, nil
}

func (l LazyDataBytes) AsMap() (map[string]any, error) {
	var data map[string]any
	err := json.Unmarshal(l, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (l LazyDataBytes) AsJson(target any) error {
	err := json.Unmarshal(l, target)
	if err != nil {
		return err
	}

	return nil
}

func (l LazyDataBytes) GetFieldAsInt(path string) int {
	return int(gjson.GetBytes(l, path).Int())
}

func (l LazyDataBytes) GetFieldAsFloat(path string) float64 {
	return gjson.GetBytes(l, path).Float()
}

func (l LazyDataBytes) GetFieldAsString(path string) string {
	return gjson.GetBytes(l, path).String()
}

func (l LazyDataBytes) GetFieldAsBool(path string) bool {
	return gjson.GetBytes(l, path).Bool()
}
