package cloudy

import "time"

type User struct {
	ID         int64      `db:"id,omitempty" json:"id"`
	Fullname   string     `db:"fullname" json:"fullname"`
	Email      string     `db:"email" json:"email"`
	Password   string     `db:"password" json:"-"`
	UType      string     `db:"utype" json:"utype"`
	IsVerified bool       `db:"is_verified" json:"is_verified"`
	IsDisabled bool       `db:"is_disabled" json:"is_disabled"`
	CreatedAt  *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt  *time.Time `db:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type Team struct {
	ID          int64      `db:"id,omitempty" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	OwnerID     int64      `db:"owner_id" json:"owner_id"`
	IsDeleted   bool       `db:"is_deleted" json:"is_deleted"`
	CreatedAt   *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt   *time.Time `db:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type TeamMember struct {
	ID        int64      `db:"id,omitempty" json:"id"`
	TeamID    int64      `db:"team_id" json:"team_id"`
	UserID    int64      `db:"user_id" json:"user_id"`
	IsDeleted bool       `db:"is_deleted" json:"is_deleted"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type TeamPotatoInstance struct {
	ID          int64      `db:"id,omitempty" json:"id"`
	Slug        string     `db:"slug" json:"slug"`
	Description string     `db:"description" json:"description"`
	TeamID      int64      `db:"team_id" json:"team_id"`
	IsDeleted   bool       `db:"is_deleted" json:"is_deleted"`
	CreatedAt   *time.Time `db:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt   *time.Time `db:"updated_at,omitempty" json:"updated_at,omitempty"`
}
