package repohub

type HttpPackageV1 struct {
	Name        string `json:"name" toml:"name"`
	Slug        string `json:"slug" toml:"slug"`
	Info        string `json:"info" toml:"info"`
	DownloadUrl string `json:"download_url" toml:"download_url"`
}

type HttpPackageIndexV1 []HttpPackageV1
