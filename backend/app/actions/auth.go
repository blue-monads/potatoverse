package actions

import (
	"errors"
	"regexp"
	"strings"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
)

type LoginOpts struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

	user.Password = ""
	user.ExtraMeta = ""

	return &LoginResponse{
		AccessToken:    token,
		UserInfo:       user,
		PortalPageType: "login",
	}, nil

}
