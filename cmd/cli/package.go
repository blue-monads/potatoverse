package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/pelletier/go-toml/v2"
)

func (c *PackageBuildCmd) doRun(_ *kong.Context) error {
	potatoTomlFile, err := os.ReadFile(c.PotatoTomlFile)
	if err != nil {
		return err
	}

	potatoToml := models.PotatoPackage{}
	err = toml.Unmarshal(potatoTomlFile, &potatoToml)
	if err != nil {
		return err
	}

	if c.OutputZipFile == "" {
		c.OutputZipFile = fmt.Sprintf("%s.zip", potatoToml.Slug)
	}

	zipFile, err := os.Create(c.OutputZipFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	potatoFileDir := path.Dir(c.PotatoTomlFile)

	err = includeSubFolder(potatoFileDir, potatoFileDir, potatoToml.FilesDir, zipWriter)
	if err != nil {
		return err
	}

	potatoToml.FilesDir = ""
	potatoToml.DevToken = ""

	potatoJson, err := json.Marshal(potatoToml)
	if err != nil {
		return err
	}

	pfile, err := zipWriter.Create("potato.json")
	if err != nil {
		return err
	}
	_, err = pfile.Write(potatoJson)
	if err != nil {
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	fmt.Printf("Package built successfully: %s\n", c.OutputZipFile)

	return nil
}

func includeSubFolder(basePath, folder, name string, zipWriter *zip.Writer) error {

	fullPath := path.Join(folder, name)

	files, err := os.ReadDir(fullPath)
	if err != nil {
		return err
	}

	for _, file := range files {

		if file.IsDir() {
			err = includeSubFolder(basePath, fullPath, file.Name(), zipWriter)
			if err != nil {
				return err
			}
			continue
		}

		// Create the relative path for this file within the zip
		filePath := path.Join(fullPath, file.Name())
		targetPath := strings.TrimPrefix(filePath, basePath)
		targetPath = strings.TrimPrefix(targetPath, "/")

		zfile, err := zipWriter.Create(targetPath)
		if err != nil {
			return err
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		_, err = zfile.Write(fileData)
		if err != nil {
			return err
		}

	}

	return nil
}
