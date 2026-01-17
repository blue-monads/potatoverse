package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/alecthomas/kong"
	"github.com/blue-monads/potatoverse/cmd/cli/pkgutils"
)

func (c *PackageBuildCmd) Run(_ *kong.Context) error {

	err := RunBuildCommand(c.PotatoTomlFile)
	if err != nil {
		return err
	}

	_, err = PackageFiles(c.PotatoTomlFile, c.OutputZipFile)
	if err != nil {
		return err
	}

	return nil
}

// simple.chip.zip
// simple.czip

func RunBuildCommand(potatoTomlFile string) error {
	fmt.Printf("Running build command\n")

	potatoToml, err := pkgutils.ReadPotatoToml(potatoTomlFile)
	if err != nil {
		return err
	}

	buildCommand := ""

	if potatoToml.Developer != nil &&
		potatoToml.Developer.BuildCommand != "" {
		buildCommand = potatoToml.Developer.BuildCommand
	}

	if buildCommand == "" {
		fmt.Println("No build command found, skipping build")
		return nil
	}

	cmd := exec.Command("bash", "-c", buildCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Build command failed: %s\n", err)
		return err
	}

	fmt.Printf("Build command completed successfully\n")
	return nil
}

func PackageFiles(potatoTomlFile string, outputZipFile string) (string, error) {
	fmt.Printf("Package files start\n")

	potatoToml, err := pkgutils.ReadPotatoToml(potatoTomlFile)
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

	err = pkgutils.PackageFilesV2(potatoFileDir, potatoToml.Developer, zipWriter)
	if err != nil {
		return "", err
	}

	potatoToml.Developer = nil

	potatoMap, err := pkgutils.ReadPotatoMap(potatoTomlFile)
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
