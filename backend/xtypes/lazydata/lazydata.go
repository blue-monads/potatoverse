package lazydata

import "encoding/json"

type LazyData interface {
	AsBytes() ([]byte, error)

	AsMap() (map[string]any, error)
	// AsJSON struct target
	AsJson(target any) error
}

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
