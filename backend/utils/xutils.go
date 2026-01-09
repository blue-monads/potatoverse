package xutils

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"

	"github.com/blue-monads/potatoverse/backend/xtypes/models"
)

func GetPackageManifest(zipFile string) ([]byte, error) {
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

			data, err := io.ReadAll(jsonFile)
			if err != nil {
				return data, nil
			}

			return data, nil

		}
	}

	return nil, errors.New("potato.json not found")
}

func ReadPackageManifestFromZip(zipFile string) (*models.PotatoPackage, error) {

	data, err := GetPackageManifest(zipFile)
	if err != nil {
		return nil, err
	}

	pkg := &models.PotatoPackage{}
	err = json.Unmarshal(data, pkg)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}
