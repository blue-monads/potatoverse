package xutils

import (
	"archive/zip"
	"encoding/json"
	"errors"

	"github.com/blue-monads/turnix/backend/xtypes/models"
)

func ReadPackageManifestFromZip(zipFile string) (*models.PotatoPackage, error) {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		if file.Name == "potato.json" {
			jsonFile, err := file.Open()
			if err != nil {
				return nil, err
			}

			pkg := &models.PotatoPackage{}
			json.NewDecoder(jsonFile).Decode(&pkg)
			if err != nil {
				return nil, err
			}

			return pkg, nil
		}
	}

	return nil, errors.New("potato.json not found")
}
