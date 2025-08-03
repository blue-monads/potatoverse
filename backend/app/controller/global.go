package controller

import (
	"encoding/json"
	"fmt"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
)

type AppFingerPrint struct {
	Version          string `json:"version"`
	Commit           string `json:"commit"`
	BuildAt          string `json:"build_at"`
	MasterSecretHash string `json:"master_secret_hash"`
}

func (c *Controller) HasFingerprint() (bool, error) {

	config, err := c.database.GetGlobalConfig("fingerprint", "CORE")
	if err != nil {
		return false, err
	}

	if config == nil || config.Value == "" {
		has, err := c.database.HasTable("GlobalConfig")
		if err != nil {
			return false, err
		}

		if !has {
			return false, nil
		}

		return false, fmt.Errorf("Unknown error: fingerprint not found in global config")
	}

	return true, nil
}

func (c *Controller) GetAppFingerPrint() (*AppFingerPrint, error) {
	config, err := c.database.GetGlobalConfig("fingerprint", "CORE")
	if err != nil {
		return nil, err
	}

	fingerPrint := &AppFingerPrint{}

	err = json.Unmarshal([]byte(config.Value), &fingerPrint)
	if err != nil {
		return nil, err
	}

	return fingerPrint, nil
}

func (c *Controller) SetAppFingerPrint(fingerPrint *AppFingerPrint) error {
	data, err := json.Marshal(fingerPrint)
	if err != nil {
		return err
	}

	config := &models.GlobalConfig{
		Key:       "fingerprint",
		GroupName: "CORE",
		Value:     string(data),
	}

	_, err = c.database.AddGlobalConfig(config)
	if err != nil {
		return err
	}

	return nil
}
