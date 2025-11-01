package dbmodels

type SpaceCapability struct {
	ID             int64  `json:"id" db:"id,omitempty"`
	Name           string `json:"name" db:"name"`
	CapabilityType string `json:"capability_type" db:"capability_type"`
	InstallID      int64  `json:"install_id" db:"install_id"`
	SpaceID        int64  `json:"space_id" db:"space_id"`
	Options        string `json:"options" db:"options"`
	ExtraMeta      string `json:"extrameta" db:"extrameta,omitempty"`
}
