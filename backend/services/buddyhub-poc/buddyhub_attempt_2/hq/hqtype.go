package hq

type SelfInfo struct {
	Address []AddressInfo `json:"address"`
	Port    int           `json:"port"`
	Pubkey  string        `json:"pubkey"`
}

type AddressInfo struct {
	Type string `json:"type"`
	Addr string `json:"addr"`
}
