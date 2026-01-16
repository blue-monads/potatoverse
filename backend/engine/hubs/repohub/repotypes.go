package repohub

type PotatoPackage struct {
	Name          string   `json:"name" toml:"name"`
	Slug          string   `json:"slug" toml:"slug"`
	Info          string   `json:"info" toml:"info"`
	Tags          []string `json:"tags" toml:"tags"`
	FormatVersion string   `json:"format_version" toml:"format_version"`
	AuthorName    string   `json:"author_name" toml:"author_name"`
	AuthorEmail   string   `json:"author_email" toml:"author_email"`
	AuthorSite    string   `json:"author_site" toml:"author_site"`
	SourceCode    string   `json:"source_code" toml:"source_code"`
	License       string   `json:"license" toml:"license"`
	Version       string   `json:"version" toml:"version"`
	Versions      []string `json:"versions" toml:"versions"`
}
