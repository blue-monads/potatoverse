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
	CreatedAt      time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at,omitempty"`
}
