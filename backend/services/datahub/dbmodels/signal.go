package dbmodels

import "time"

type Signal struct {
	ID                int64      `json:"id" db:"id,omitempty"`
	SignalKey         string     `json:"signal_key" db:"signal_key"`
	EmitterInstallID  int64      `json:"emitter_install_id" db:"emitter_install_id"`
	EmitterSpaceID    int64      `json:"emitter_space_id" db:"emitter_space_id"`
	ReceiverInstallID int64      `json:"receiver_install_id" db:"receiver_install_id"`
	ReceiverSpaceID   int64      `json:"receiver_space_id" db:"receiver_space_id"`
	ReceiverHandler   string     `json:"receiver_handler" db:"receiver_handler"`
	ManagedBy         string     `json:"managed_by" db:"managed_by"` // emitter, receiver, both
	ExpiresOn         int64      `json:"expires_on" db:"expires_on"`
	MaxRetries        int64      `json:"max_retries" db:"max_retries"`
	RetryDelay        int64      `json:"retry_delay" db:"retry_delay"`
	CreatedBy         int64      `json:"created_by" db:"created_by"`
	Disabled          bool       `json:"disabled" db:"disabled"`
	CreatedAt         *time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt         *time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

type SignalLite struct {
	ID                int64  `json:"id" db:"id,omitempty"`
	SignalKey         string `json:"signal_key" db:"signal_key"`
	EmitterInstallID  int64  `json:"emitter_install_id" db:"emitter_install_id"`
	EmitterSpaceID    int64  `json:"emitter_space_id" db:"emitter_space_id"`
	ReceiverInstallID int64  `json:"receiver_install_id" db:"receiver_install_id"`
	ReceiverSpaceID   int64  `json:"receiver_space_id" db:"receiver_space_id"`
	ReceiverHandler   string `json:"receiver_handler" db:"receiver_handler"`
	MaxRetries        int64  `json:"max_retries" db:"max_retries"`
	RetryDelay        int64  `json:"retry_delay" db:"retry_delay"`
	ExpiresOn         int64  `json:"expires_on" db:"expires_on"`
	Disabled          bool   `json:"disabled" db:"disabled"`
}

type SignalEvent struct {
	ID            int64      `json:"id" db:"id,omitempty"`
	SignalID      int64      `json:"signal_id" db:"signal_id"`
	Payload       []byte     `json:"payload" db:"payload"`
	Metadata      string     `json:"metadata" db:"metadata"`
	Status        string     `json:"status" db:"status"` // new, scheduled, processed, failed, delayed, expired
	DelayedUntil  int64      `json:"delayed_until" db:"delayed_until"`
	RetryCount    int64      `json:"retry_count" db:"retry_count"`
	LastRetriedAt int64      `json:"last_retried_at" db:"last_retried_at"`
	Error         string     `json:"error" db:"error"`
	ExtraMeta     string     `json:"extrameta" db:"extrameta"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at,omitempty"`
}
