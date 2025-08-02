package backend

type DomainInfo struct {
	Domain string // example.com
	Mode   string // shared(example.com), isolated(*.example.com), etc.
}

type ResourceProvisiner interface {
	GetAppDomain() DomainInfo
	GetSpaceDomain() string // space-11.example.com

}
