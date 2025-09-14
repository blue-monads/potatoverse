package actions

import "github.com/blue-monads/turnix/backend/engine"

func (c *Controller) ListEPackages() ([]engine.EPackage, error) {
	return engine.ListEPackages()
}
