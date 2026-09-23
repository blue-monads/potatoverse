package actions

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func (c *Controller) CreateSignal(installId int64, data *dbmodels.Signal) (*dbmodels.Signal, error) {
	if data.EmitterInstallID == 0 {
		data.EmitterInstallID = installId
	}

	id, err := c.database.GetSignalOps().AddSignal(data)
	if err != nil {
		return nil, err
	}

	c.engine.RefreshSignalIndex()

	return c.database.GetSignalOps().GetSignal(id)
}

func (c *Controller) UpdateSignalByID(installId int64, signalId int64, data map[string]any) (*dbmodels.Signal, error) {
	_, err := c.GetSignalByID(installId, signalId)
	if err != nil {
		return nil, err
	}

	err = c.database.GetSignalOps().UpdateSignal(signalId, data)
	if err != nil {
		return nil, err
	}

	c.engine.RefreshSignalIndex()

	return c.database.GetSignalOps().GetSignal(signalId)
}

func (c *Controller) DeleteSignalByID(installId int64, signalId int64) error {
	err := c.database.GetSignalOps().RemoveSignal(signalId)
	if err != nil {
		return err
	}

	c.engine.RefreshSignalIndex()

	return nil
}

func (c *Controller) QuerySignals(installId int64, spaceId int64, cond map[any]any) ([]dbmodels.Signal, error) {
	return c.database.GetSignalOps().QuerySignals(installId, spaceId, cond)
}

func (c *Controller) GetSignalByID(installId int64, signalId int64) (*dbmodels.Signal, error) {
	return c.database.GetSignalOps().GetSignal(signalId)
}

func (c *Controller) QuerySignalEvents(installId int64, signalId int64, status string, limit, offset int64) ([]dbmodels.SignalEvent, error) {
	return c.database.GetSignalOps().QuerySignalEvents(installId, signalId, status, limit, offset)
}

func (c *Controller) EmitSignal(installId int64, opts *xtypes.SignalOptions) error {
	if opts.EmitterInstallId == 0 {
		opts.EmitterInstallId = installId
	}
	return c.engine.PublishSignal(opts)
}
