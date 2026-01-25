package hq

type SelfInfo struct {
	Address []AddressInfo
	Port    int
	Pubkey  string
}

type AddressInfo struct {
	AddrType string
	Addr     string
}
