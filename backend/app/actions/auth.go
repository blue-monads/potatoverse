package actions

import (
	"errors"
	"regexp"
	"strings"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/rs/xid"
)

type LoginOpts struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

func (c *Controller) Login(opts LoginOpts) (string, error) {

	var user *models.User

	if strings.Contains(opts.Email, "@") {
		_user, err := c.database.GetUserByEmail(opts.Email)
		if err != nil {
			return "", err
		}
		user = _user
	} else if phoneRegex.MatchString(opts.Email) {
		return "", errors.New("implement login by phone")
	} else {
		return "", errors.New("implement login by username")
	}

	// fixme => hash it
	if user.Password != opts.Password {
		return "", errors.New("invalid password")
	}

	token, err := c.signer.SignAccess(&signer.AccessClaim{
		XID:    xid.New().String(),
		UserId: int64(user.ID),
	})

	if err != nil {
		return "", err
	}

	return token, nil

}
