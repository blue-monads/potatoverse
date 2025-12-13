package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/alecthomas/kong"
)

func (c *PackageBuildCmd) Run(_ *kong.Context) error {
	_, err := PackageFiles(c.PotatoTomlFile, c.OutputZipFile)
	if err != nil {
		return err
	}

	return nil
}

// simple.chip.zip
// simple.czip

func PackageFiles(potatoTomlFile string, outputZipFile string) (string, error) {
	fmt.Printf("PackageFiles start\n")

	potatoToml, err := readPotatoToml(potatoTomlFile)
	if err != nil {
		return "", err
	}

	if outputZipFile == "" {
		outputZipFile = potatoToml.Developer.OutputZipFile
		if outputZipFile == "" {
			outputZipFile = fmt.Sprintf("%s.zip", potatoToml.Slug)
		}
	}

	zipFile, err := os.Create(outputZipFile)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	potatoFileDir := path.Dir(potatoTomlFile)

	err = packageFilesV2(potatoFileDir, potatoToml.Developer, zipWriter)
	if err != nil {
		return "", err
	}

	potatoToml.Developer = nil

	potatoMap, err := readPotatoMap(potatoTomlFile)
	if err != nil {
		return "", err
	}

	pfile, err := zipWriter.Create("potato.json")
	if err != nil {
		return "", err
	}

	potatoJson, err := json.Marshal(potatoMap)
	if err != nil {
		return "", err
	}

	_, err = pfile.Write(potatoJson)
	if err != nil {
		return "", err
	}

	err = zipWriter.Close()
	if err != nil {
		return "", err
	}

	fmt.Printf("Package built successfully: %s\n", outputZipFile)

	return outputZipFile, nil
}
