package dbmodels

import "time"

type UserGroup struct {
	Name      string     `json:"name" db:"name"`
	Info      string     `json:"info" db:"info"`
	CreatedAt *time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

type User struct {
	ID              int64      `json:"id" db:"id,omitempty"`
	Name            string     `json:"name" db:"name"`
	Utype           string     `json:"utype" db:"utype,omitempty"`
	Ugroup          string     `json:"ugroup" db:"ugroup,omitempty"`
	Email           string     `json:"email" db:"email,omitempty"`
	Phone           *string    `json:"phone" db:"phone,omitempty"`
	Username        *string    `json:"username" db:"username,omitempty"`
	Bio             string     `json:"bio" db:"bio"`
	Password        string     `json:"password" db:"password,omitempty"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	ExtraMeta       string     `json:"extrameta" db:"extrameta,omitempty"`
	CreatedAt       *time.Time `json:"createdAt" db:"created_at,omitempty"`
	OwnerUserId     int64      `json:"owner_user_id" db:"owner_user_id,omitempty"`
	OwnerSpaceId    int64      `json:"owner_space_id" db:"owner_space_id,omitempty"`
	MessageReadHead int64      `json:"msg_read_head" db:"msg_read_head,omitempty"`
	Disabled        bool       `json:"disabled" db:"disabled,omitempty"`
	IsDeleted       bool       `json:"is_deleted" db:"is_deleted,omitempty"`
}

type UserMessage struct {
	ID            int64      `json:"id" db:"id,omitempty"`
	Name          string     `json:"title" db:"title"`
	Type          string     `json:"type" db:"type"`
	Contents      string     `json:"contents" db:"contents"`
	ToUser        int64      `json:"to_user" db:"to_user"`
	FromUser      int64      `json:"from_user" db:"from_user_id"`
	FromProject   int64      `json:"from_project" db:"from_project_id"`
	IsRead        bool       `json:"is_read" db:"is_read,omitempty"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at,omitempty"`
	CallbackToken string     `json:"callback_token" db:"callback_token,omitempty"`
	Disabled      bool       `json:"disabled" db:"disabled,omitempty"`
}

type UserDevice struct {
	ID        int64      `json:"id" db:"id,omitempty"`
	Name      string     `json:"name" db:"name"`
	Dtype     string     `json:"dtype" db:"dtype"`
	TokenHash string     `json:"token_hash" db:"token_hash"`
	UserId    int64      `json:"user_id" db:"user_id"`
	ProjectId int64      `json:"project_id" db:"project_id,omitempty"`
	LastIp    string     `json:"last_ip" db:"last_ip"`
	LastLogin string     `json:"last_login" db:"last_login"`
	ExtraMeta string     `json:"extrameta" db:"extrameta,omitempty"`
	ExpiresOn *time.Time `json:"expires_on" db:"expires_on,omitempty"`
	CreatedAt *time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

type UserInvite struct {
	ID            int64      `json:"id" db:"id,omitempty"`
	Email         string     `json:"email" db:"email"`
	Role          string     `json:"role" db:"role"`
	Status        string     `json:"status" db:"status"` // pending, accepted, rejected
	InvitedBy     int64      `json:"invited_by" db:"invited_by"`
	InvitedAsType string     `json:"invited_as_type" db:"invited_as_type"` // user, admin, moderator, developer
	ExpiresOn     *time.Time `json:"expires_on" db:"expires_on,omitempty"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at,omitempty"`
}
