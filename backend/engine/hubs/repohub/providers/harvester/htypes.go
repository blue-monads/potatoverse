package harvester

import "github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/repotypes"

type PotatoField struct {
	Name               string                    `json:"name"`
	Info               string                    `json:"info"`
	Type               string                    `json:"type"`
	ZipTemplate        string                    `json:"zip_template"`
	IndexedTags        []string                  `json:"indexed_tags"`
	IndexedTagTemplate string                    `json:"indexed_tag_template"`
	Potatoes           []repotypes.PotatoPackage `json:"potatoes"`
}
