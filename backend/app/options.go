package app

type Options struct {
	Name         string `json:"name"`
	Port         int    `json:"port"`
	Host         string `json:"host"`
	MasterSecret string `json:"master_secret"`
	Debug        bool   `json:"debug"`
	WorkingDir   string `json:"working_dir"`
}
