package main

type BPrint struct {
	Name            string `json:"name"`
	RunTimeVersion  string `json:"runtime_version"`
	Executor        string `json:"executor"`
	ExecutorVersion string `json:"executor_version"`
	HomePage        string `json:"home_page"`
	Logo            string `json:"logo"`
	SingleTon       bool   `json:"single_ton"`
}

type DomainInfo struct {
	Domain string // example.com
	Mode   string // shared(example.com), isolated(*.example.com), etc.
}

type ResourceProvisiner interface {
	GetAppDomain() DomainInfo
	GetSpaceDomain() string // space-11.example.com

}
