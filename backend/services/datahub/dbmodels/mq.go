package dbmodels

import "time"

type MQEvent struct {
	ID        int64     `json:"id" db:"id,omitempty"`
	InstallID int64     `json:"install_id" db:"install_id"`
	Name      string    `json:"name" db:"name"`
	Payload   []byte    `json:"payload" db:"payload"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

type MQEventTarget struct {
	ID             int64     `json:"id" db:"id,omitempty"`
	EventID        int64     `json:"event_id" db:"event_id"`
	SubscriptionID int64     `json:"subscription_id" db:"subscription_id"`
	Status         string    `json:"status" db:"status"`
	DelayedUntil   int64     `json:"delayed_until" db:"delayed_until"`
	RetryCount     int64     `json:"retry_count" db:"retry_count"`
	CreatedAt      time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

type MQSubscription struct {
	ID             int64  `json:"id" db:"id,omitempty"`
	InstallID      int64  `json:"install_id" db:"install_id"`
	SpaceID        int64  `json:"space_id" db:"space_id"`
	EventKey       string `json:"event_key" db:"event_key"`
	TargetType     string `json:"target_type" db:"target_type"` // push, email, sms, webhook, script
	TargetEndpoint string `json:"target_endpoint" db:"target_endpoint"`
	TargetOptions  string `json:"target_options" db:"target_options"` // JSON
	TargetCode     string `json:"target_code" db:"target_code"`
	Rules          string `json:"rules" db:"rules"`         // JSON
	Transform      string `json:"transform" db:"transform"` // JSON
	DelayStart     int64  `json:"delay_start" db:"delay_start"`
	RetryDelay     int64  `json:"retry_delay" db:"retry_delay"`
	MaxRetries     int64  `json:"max_retries" db:"max_retries"`
	TargetSpaceID  int64  `json:"target_space_id" db:"target_space_id"`

	ExtraMeta string     `json:"extrameta" db:"extrameta,omitempty"` // JSON
	CreatedBy int64      `json:"created_by" db:"created_by"`
	Disabled  bool       `json:"disabled" db:"disabled"`
	CreatedAt *time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

type MQSubscriptionLite struct {
	ID        int64  `json:"id" db:"id,omitempty"`
	InstallID int64  `json:"install_id" db:"install_id"`
	SpaceID   int64  `json:"space_id" db:"space_id"`
	EventKey  string `json:"event_key" db:"event_key"`
}
