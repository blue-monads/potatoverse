package core

type PresignedOptions struct {
	Uid      int64  `json:"uid,omitempty"`
	Path     string `json:"path,omitempty"`
	FileName string `json:"file_name,omitempty"`
	Expiry   int64  `json:"expiry,omitempty"`
}
