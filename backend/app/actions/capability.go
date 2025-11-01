package actions

import (
	"encoding/json"
	"errors"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) CreateSpaceCapability(installId int64, data map[string]any) (*dbmodels.SpaceCapability, error) {
	// Validate required fields
	name, ok := data["name"].(string)
	if !ok || name == "" {
		return nil, errors.New("name is required")
	}

	capabilityType, ok := data["capability_type"].(string)
	if !ok || capabilityType == "" {
		return nil, errors.New("capability_type is required")
	}

	spaceId, ok := data["space_id"].(float64)
	if !ok {
		return nil, errors.New("space_id is required")
	}

	// Extract options JSON
	var optionsJSON string
	if options, ok := data["options"]; ok {
		if optionsStr, ok := options.(string); ok {
			optionsJSON = optionsStr
		} else {
			// If it's a map, marshal it to JSON
			optsBytes, err := json.Marshal(options)
			if err != nil {
				return nil, errors.New("invalid options format")
			}
			optionsJSON = string(optsBytes)
		}
	} else {
		optionsJSON = "{}"
	}

	// Extract extrameta JSON
	var extraMetaJSON string
	if extrameta, ok := data["extrameta"]; ok {
		if extrametaStr, ok := extrameta.(string); ok {
			extraMetaJSON = extrametaStr
		} else {
			// If it's a map, marshal it to JSON
			metaBytes, err := json.Marshal(extrameta)
			if err != nil {
				return nil, errors.New("invalid extrameta format")
			}
			extraMetaJSON = string(metaBytes)
		}
	} else {
		extraMetaJSON = "{}"
	}

	capability := &dbmodels.SpaceCapability{
		Name:           name,
		CapabilityType: capabilityType,
		SpaceID:        int64(spaceId),
		Options:        optionsJSON,
		ExtraMeta:      extraMetaJSON,
	}

	err := c.database.GetSpaceOps().AddSpaceCapability(installId, capability)
	if err != nil {
		return nil, err
	}

	// Get the created capability
	return c.database.GetSpaceOps().GetSpaceCapability(installId, name)
}

func (c *Controller) UpdateSpaceCapabilityByID(installId int64, capabilityId int64, data map[string]any) (*dbmodels.SpaceCapability, error) {
	// Get the existing capability
	capability, err := c.GetSpaceCapabilityByID(installId, capabilityId)
	if err != nil {
		return nil, err
	}

	// Handle options if present
	if options, ok := data["options"]; ok {
		if optionsStr, ok := options.(string); ok {
			data["options"] = optionsStr
		} else {
			// If it's a map, marshal it to JSON
			optsBytes, err := json.Marshal(options)
			if err != nil {
				return nil, errors.New("invalid options format")
			}
			data["options"] = string(optsBytes)
		}
	}

	// Handle extrameta if present
	if extrameta, ok := data["extrameta"]; ok {
		if extrametaStr, ok := extrameta.(string); ok {
			data["extrameta"] = extrametaStr
		} else {
			// If it's a map, marshal it to JSON
			metaBytes, err := json.Marshal(extrameta)
			if err != nil {
				return nil, errors.New("invalid extrameta format")
			}
			data["extrameta"] = string(metaBytes)
		}
	}

	// Update using installId and name
	err = c.database.GetSpaceOps().UpdateSpaceCapability(installId, capability.Name, data)
	if err != nil {
		return nil, err
	}

	// Return updated entry
	return c.database.GetSpaceOps().GetSpaceCapability(installId, capability.Name)
}

func (c *Controller) DeleteSpaceCapabilityByID(installId int64, capabilityId int64) error {
	// Get the existing capability to find name
	capability, err := c.GetSpaceCapabilityByID(installId, capabilityId)
	if err != nil {
		return err
	}

	return c.database.GetSpaceOps().RemoveSpaceCapability(installId, capability.Name)
}

func (c *Controller) QuerySpaceCapabilities(installId int64, cond map[any]any) ([]dbmodels.SpaceCapability, error) {
	return c.database.GetSpaceOps().QuerySpaceCapabilities(installId, cond)
}

func (c *Controller) GetSpaceCapabilityByID(installId int64, capabilityId int64) (*dbmodels.SpaceCapability, error) {
	capability, err := c.database.GetSpaceOps().GetSpaceCapabilityByID(installId, capabilityId)
	if err != nil {
		return nil, err
	}

	return capability, nil
}
