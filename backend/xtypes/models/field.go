package models

type PotatoField struct {
	Name        string        `json:"name"`
	Info        string        `json:"info"`
	Slug        string        `json:"slug"`
	Tags        []string      `json:"tags"`
	ZipTemplate string        `json:"zip_template"`
	Potatoes    []FieldPotato `json:"potatoes"`
}

type FieldPotato struct {
	Name          string   `json:"name"`
	Info          string   `json:"info"`
	Slug          string   `json:"slug"`
	Tags          []string `json:"tags"`
	FormatVersion string   `json:"format_version"`
	AuthorName    string   `json:"author_name"`
	AuthorEmail   string   `json:"author_email"`
	AuthorSite    string   `json:"author_site"`
	License       string   `json:"license"`
	Versions      []string `json:"versions"`
}
