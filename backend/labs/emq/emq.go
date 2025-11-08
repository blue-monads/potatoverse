package emq

type MQEventLite struct {
	ID        int64 `json:"id"`
	InstallID int64 `json:"install_id"`
}

type MQSynk interface {
	Poll() ([]MQEventLite, error)
	AddEvent(installId int64, name string, payload []byte) error
	GetEvent(id int64) ([]byte, error)
	UpdateEvent(id int64, data map[string]any) error
}
