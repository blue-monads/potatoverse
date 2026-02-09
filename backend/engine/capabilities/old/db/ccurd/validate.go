package ccurd

import "regexp"

type Methods struct {
	Table        string                     `json:"table"`
	Mode         string                     `json:"mode"`
	EventName    string                     `json:"event_name"`
	Validators   map[string]*ValidationItem `json:"validators"`
	StaticFields map[string]any             `json:"static_fields"`
}

type ValidationItem struct {
	Type          string         `json:"type"`
	Required      bool           `json:"required"`
	Min           int64          `json:"min"`
	Max           int64          `json:"max"`
	Regex         string         `json:"regex"`
	compiledRegex *regexp.Regexp `json:"-"`
}
