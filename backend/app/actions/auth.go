package actions

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
)

type LoginOpts struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	OldToken   string `json:"old_token"`
	ClientIP   string `json:"-"` // set by server from request
	DeviceName string `json:"device_name"`
}

type LoginResponse struct {
	AccessToken    string         `json:"access_token"`
	UserInfo       *dbmodels.User `json:"user_info"`
	PortalPageType string         `json:"portal_page_type"`
}

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

func (c *Controller) Login(opts *LoginOpts) (*LoginResponse, error) {

	var user *dbmodels.User

	if strings.Contains(opts.Username, "@") {
		_user, err := c.database.GetUserOps().GetUserByEmail(opts.Username)
		if err != nil {
			return nil, err
		}
		user = _user
	} else if phoneRegex.MatchString(opts.Username) {
		return nil, errors.New("implement login by phone")
	} else {
		return nil, errors.New("implement login by username")
	}

	// fixme => hash it
	if user.Password != opts.Password {
		return nil, errors.New("invalid password")
	}

	token, err := c.signer.SignAccess(&signer.AccessClaim{
		UserId: int64(user.ID),
	})
	if err != nil {
		return nil, err
	}

	// Create or update device record for this session
	tokenHash := HashToken(token)
	now := time.Now()
	expiresOn := now.Add(30 * 24 * time.Hour)
	lastLogin := now.Format(time.RFC3339)
	userOps := c.database.GetUserOps()

	if opts.OldToken != "" {
		oldHash := HashToken(opts.OldToken)
		existing, err := userOps.GetUserDeviceByTokenHash(int64(user.ID), oldHash)
		if err == nil && existing != nil {
			_ = userOps.UpdateUserDevice(existing.ID, map[string]any{
				"token_hash": tokenHash,
				"last_ip":    opts.ClientIP,
				"last_login": lastLogin,
				"expires_on": expiresOn,
				"updated_at": now,
			})
			user.Password = ""
			user.ExtraMeta = ""
			return &LoginResponse{
				AccessToken:    token,
				UserInfo:       user,
				PortalPageType: "login",
			}, nil
		}
	}

	deviceName := opts.DeviceName
	if deviceName == "" {
		deviceName = "Session"
	}
	_, _ = userOps.AddUserDevice(&dbmodels.UserDevice{
		Name:      deviceName,
		Dtype:     "session",
		TokenHash: tokenHash,
		UserId:    int64(user.ID),
		LastIp:    opts.ClientIP,
		LastLogin: lastLogin,
		ExtraMeta: "{}",
		ExpiresOn: &expiresOn,
	})

	user.Password = ""
	user.ExtraMeta = ""

	return &LoginResponse{
		AccessToken:    token,
		UserInfo:       user,
		PortalPageType: "login",
	}, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
