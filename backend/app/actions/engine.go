package actions

import (
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/datahub/models"
)

func (c *Controller) ListEPackages() ([]engine.EPackage, error) {
	return engine.ListEPackages()
}

func (c *Controller) ListInstalledSpaces(userId int64) ([]models.Space, error) {

	spaces, err := c.database.ListSpaces()
	if err != nil {
		return nil, err
	}

	tpSpaces, err := c.database.ListThirdPartySpaces(userId, "")
	if err != nil {
		return nil, err
	}

	spaces = append(spaces, tpSpaces...)

	return spaces, nil

}
