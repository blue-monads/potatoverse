package hq

type SelfInfo struct {
	Address []AddressInfo `json:"address"`
	Port    int           `json:"port"`
	Pubkey  string        `json:"pubkey"`
}

type AddressInfo struct {
	AddrType string `json:"addr_type"`
	Addr     string `json:"addr"`
}
