package dbmodels

import "time"

type MQEvent struct {
	ID        int64     `json:"id"`
	InstallID int64     `json:"install_id"`
	Name      string    `json:"name"`
	Payload   []byte    `json:"payload"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MQEventTarget struct {
	ID             int64     `json:"id"`
	EventID        int64     `json:"event_id"`
	SubscriptionID int64     `json:"subscription_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
